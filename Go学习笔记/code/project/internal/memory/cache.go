package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"example.com/quote-notes/internal/domain"
)

type entry struct {
	quote domain.Quote
	until time.Time
}

type Cache struct {
	mu    sync.Mutex
	items map[string]entry
	now   func() time.Time
}

func NewCache(now func() time.Time) *Cache {
	return &Cache{items: make(map[string]entry), now: now}
}

func (c *Cache) Get(ctx context.Context, symbol string) (domain.Quote, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.Quote{}, false, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[symbol]
	if !ok {
		return domain.Quote{}, false, nil
	}
	if !c.now().Before(e.until) {
		delete(c.items, symbol)
		return domain.Quote{}, false, nil
	}
	return e.quote, true, nil
}

func (c *Cache) Set(ctx context.Context, q domain.Quote, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ttl <= 0 {
		return errors.New("cache TTL must be positive")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[q.Symbol] = entry{quote: q, until: c.now().Add(ttl)}
	return nil
}
