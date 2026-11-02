package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/quote-notes/internal/domain"
	"example.com/quote-notes/internal/memory"
)

type fakeExchange struct {
	calls   atomic.Int64
	gate    <-chan struct{}
	entered chan struct{}
	once    sync.Once
}

func (f *fakeExchange) Fetch(ctx context.Context, symbol string) (domain.Quote, error) {
	f.calls.Add(1)
	if f.entered != nil {
		f.once.Do(func() { close(f.entered) })
	}
	if f.gate != nil {
		select {
		case <-f.gate:
		case <-ctx.Done():
			return domain.Quote{}, ctx.Err()
		}
	}
	return domain.Quote{ID: "one", Exchange: "demo", Symbol: symbol, Price: "10.00", ObservedAt: time.Unix(1, 0).UTC()}, nil
}

func TestConcurrentMissesShareCachedResult(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	f := &fakeExchange{gate: release, entered: entered}
	s := New(f, memory.NewCache(time.Now), memory.NewHistory(10), time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q, _, err := s.Latest(context.Background(), "BTCUSDT")
			if err != nil || q.Price != "10.00" {
				t.Errorf("bad result: %v", err)
			}
		}()
	}
	<-entered
	close(release)
	wg.Wait()
	if f.calls.Load() != 1 {
		t.Fatalf("upstream calls=%d", f.calls.Load())
	}
}

func TestWaitingRefreshCanBeCancelled(t *testing.T) {
	release := make(chan struct{})
	entered := make(chan struct{})
	f := &fakeExchange{gate: release, entered: entered}
	s := New(f, memory.NewCache(time.Now), memory.NewHistory(10), time.Minute)
	done := make(chan error, 1)
	go func() { _, err := s.Refresh(context.Background(), "BTCUSDT"); done <- err }()
	<-entered
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.Refresh(ctx, "ETHUSDT")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if f.calls.Load() != 1 {
		t.Fatal("cancelled request reached upstream")
	}
}

type brokenCache struct{}

func (brokenCache) Get(context.Context, string) (domain.Quote, bool, error) {
	return domain.Quote{}, false, errors.New("cache down")
}
func (brokenCache) Set(context.Context, domain.Quote, time.Duration) error {
	return errors.New("cache down")
}

func TestCacheFailureDoesNotDiscardValidQuote(t *testing.T) {
	f := &fakeExchange{}
	s := New(f, brokenCache{}, memory.NewHistory(10), time.Minute)
	q, source, err := s.Latest(context.Background(), "BTCUSDT")
	if err != nil || q.Price != "10.00" || source != "upstream" {
		t.Fatal("cache outage leaked into successful read")
	}
	if s.Snapshot().CacheErrors != 3 {
		t.Fatal("cache errors were not recorded")
	}
}

type brokenHistory struct{}

func (brokenHistory) Append(context.Context, domain.Quote) error { return errors.New("archive down") }
func (brokenHistory) List(context.Context, string, int) ([]domain.Quote, error) {
	return nil, errors.New("archive down")
}

func TestHistoryFailureDoesNotPublishCache(t *testing.T) {
	c := memory.NewCache(time.Now)
	s := New(&fakeExchange{}, c, brokenHistory{}, time.Minute)
	if _, err := s.Refresh(context.Background(), "BTCUSDT"); err == nil {
		t.Fatal("archive failure ignored")
	}
	if _, ok, _ := c.Get(context.Background(), "BTCUSDT"); ok {
		t.Fatal("unarchived observation published")
	}
}

func TestCollectorStopsOnCancelledContext(t *testing.T) {
	f := &fakeExchange{}
	s := New(f, memory.NewCache(time.Now), memory.NewHistory(10), time.Minute)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		RunCollector(ctx, s, []string{"BTCUSDT"}, time.Hour, time.Second, func(error) {})
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("collector did not stop")
	}
	if f.calls.Load() != 0 {
		t.Fatal("collector sampled after stop")
	}
}
