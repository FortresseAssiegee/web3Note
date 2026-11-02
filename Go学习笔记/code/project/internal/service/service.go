package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"example.com/quote-notes/internal/domain"
)

type Exchange interface {
	Fetch(context.Context, string) (domain.Quote, error)
}

type Cache interface {
	Get(context.Context, string) (domain.Quote, bool, error)
	Set(context.Context, domain.Quote, time.Duration) error
}

type History interface {
	Append(context.Context, domain.Quote) error
	List(context.Context, string, int) ([]domain.Quote, error)
}

type Stats struct {
	Hits          uint64 `json:"cache_hits"`
	Misses        uint64 `json:"cache_misses"`
	UpstreamCalls uint64 `json:"upstream_calls"`
	CacheErrors   uint64 `json:"cache_errors"`
}

type Service struct {
	exchange                                 Exchange
	cache                                    Cache
	history                                  History
	ttl                                      time.Duration
	gate                                     chan struct{}
	hits, misses, upstreamCalls, cacheErrors atomic.Uint64
}

func New(e Exchange, c Cache, h History, ttl time.Duration) *Service {
	if ttl <= 0 {
		panic("TTL must be positive")
	}
	return &Service{exchange: e, cache: c, history: h, ttl: ttl, gate: make(chan struct{}, 1)}
}

func (s *Service) readCache(ctx context.Context, symbol string) (domain.Quote, bool) {
	q, ok, err := s.cache.Get(ctx, symbol)
	if err != nil {
		s.cacheErrors.Add(1)
		return domain.Quote{}, false
	}
	return q, ok
}

func (s *Service) acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case s.gate <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-s.gate
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) Latest(ctx context.Context, symbol string) (domain.Quote, string, error) {
	if err := domain.ValidateSymbol(symbol); err != nil {
		return domain.Quote{}, "", err
	}
	if q, ok := s.readCache(ctx, symbol); ok {
		s.hits.Add(1)
		return q, "cache", nil
	}
	s.misses.Add(1)
	if err := s.acquire(ctx); err != nil {
		return domain.Quote{}, "", err
	}
	defer func() { <-s.gate }()
	if q, ok := s.readCache(ctx, symbol); ok {
		return q, "cache", nil
	}
	q, err := s.refreshLocked(ctx, symbol)
	return q, "upstream", err
}

func (s *Service) Refresh(ctx context.Context, symbol string) (domain.Quote, error) {
	if err := domain.ValidateSymbol(symbol); err != nil {
		return domain.Quote{}, err
	}
	if err := s.acquire(ctx); err != nil {
		return domain.Quote{}, err
	}
	defer func() { <-s.gate }()
	return s.refreshLocked(ctx, symbol)
}

func (s *Service) refreshLocked(ctx context.Context, symbol string) (domain.Quote, error) {
	s.upstreamCalls.Add(1)
	q, err := s.exchange.Fetch(ctx, symbol)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("refresh: %w", err)
	}
	if err := s.history.Append(ctx, q); err != nil {
		return domain.Quote{}, fmt.Errorf("archive observation: %w", err)
	}
	if err := s.cache.Set(ctx, q, s.ttl); err != nil {
		s.cacheErrors.Add(1)
	}
	return q, nil
}

func (s *Service) History(ctx context.Context, symbol string, limit int) ([]domain.Quote, error) {
	if err := domain.ValidateSymbol(symbol); err != nil {
		return nil, err
	}
	return s.history.List(ctx, symbol, limit)
}

func (s *Service) Snapshot() Stats {
	return Stats{s.hits.Load(), s.misses.Load(), s.upstreamCalls.Load(), s.cacheErrors.Load()}
}

func RunCollector(ctx context.Context, s *Service, symbols []string, interval, timeout time.Duration, onError func(error)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, symbol := range symbols {
				if ctx.Err() != nil {
					return
				}
				callCtx, cancel := context.WithTimeout(ctx, timeout)
				_, err := s.Refresh(callCtx, symbol)
				cancel()
				if err != nil && ctx.Err() == nil {
					onError(err)
				}
			}
		}
	}
}
