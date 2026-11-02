package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/quote-notes/internal/domain"
)

func quote(id string) domain.Quote {
	return domain.Quote{ID: id, Exchange: "demo", Symbol: "BTCUSDT", Price: "10.00", ObservedAt: time.Unix(1, 0).UTC()}
}

func TestCacheExactExpiryAndCancellation(t *testing.T) {
	now := time.Unix(10, 0)
	c := NewCache(func() time.Time { return now })
	ctx := context.Background()
	if err := c.Set(ctx, quote("a"), 2*time.Second); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := c.Get(ctx, "BTCUSDT"); err != nil || !ok {
		t.Fatal("expected hit")
	}
	now = now.Add(2 * time.Second)
	if _, ok, err := c.Get(ctx, "BTCUSDT"); err != nil || ok {
		t.Fatal("deadline must be expired")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, _, err := c.Get(cancelled, "BTCUSDT"); !errors.Is(err, context.Canceled) {
		t.Fatal("missing cancellation")
	}
	if err := c.Set(ctx, quote("b"), 0); err == nil {
		t.Fatal("zero TTL accepted")
	}
}

func TestHistoryBoundCopyAndConflict(t *testing.T) {
	h := NewHistory(2)
	ctx := context.Background()
	for _, id := range []string{"a", "b", "c", "c"} {
		if err := h.Append(ctx, quote(id)); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := h.List(ctx, "BTCUSDT", 10)
	if err != nil || len(rows) != 2 || rows[0].ID != "b" || rows[1].ID != "c" {
		t.Fatal("incorrect bounded history")
	}
	rows[0].Price = "999"
	again, _ := h.List(ctx, "BTCUSDT", 1)
	if len(again) != 1 || again[0].Price != "10.00" {
		t.Fatal("caller modified stored history")
	}
	conflict := quote("c")
	conflict.Price = "11.00"
	if h.Append(ctx, conflict) == nil {
		t.Fatal("conflicting observation accepted")
	}
	if _, err := h.List(ctx, "BTCUSDT", 101); err == nil {
		t.Fatal("unbounded query accepted")
	}
}
