package memory

import (
	"context"
	"errors"
	"sync"

	"example.com/quote-notes/internal/domain"
)

type History struct {
	mu   sync.Mutex
	rows map[string][]domain.Quote
	max  int
}

func NewHistory(maxPerSymbol int) *History {
	if maxPerSymbol < 1 {
		panic("history capacity must be positive")
	}
	return &History{rows: make(map[string][]domain.Quote), max: maxPerSymbol}
}

func (h *History) Append(ctx context.Context, q domain.Quote) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if q.ID == "" {
		return errors.New("observation ID is required")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	rows := h.rows[q.Symbol]
	for _, old := range rows {
		if old.ID == q.ID {
			if old != q {
				return errors.New("same observation ID has different content")
			}
			return nil
		}
	}
	if len(rows) == h.max {
		copy(rows, rows[1:])
		rows[len(rows)-1] = q
	} else {
		rows = append(rows, q)
	}
	h.rows[q.Symbol] = rows
	return nil
}

func (h *History) List(ctx context.Context, symbol string, limit int) ([]domain.Quote, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit < 1 || limit > 100 {
		return nil, errors.New("limit must be between 1 and 100")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	rows := h.rows[symbol]
	if len(rows) > limit {
		rows = rows[len(rows)-limit:]
	}
	out := make([]domain.Quote, len(rows))
	copy(out, rows)
	return out, nil
}
