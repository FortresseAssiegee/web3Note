package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"example.com/quote-notes/internal/domain"
)

type HTTP struct {
	Source  string
	BaseURL string
	Client  *http.Client
	Now     func() time.Time
}

func (e *HTTP) Fetch(ctx context.Context, symbol string) (domain.Quote, error) {
	if err := domain.ValidateSymbol(symbol); err != nil {
		return domain.Quote{}, err
	}
	u, err := url.Parse(e.BaseURL)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("parse upstream URL: %w", err)
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/api/v3/ticker/price"
	query := u.Query()
	query.Set("symbol", symbol)
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return domain.Quote{}, err
	}
	res, err := e.Client.Do(req)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("fetch quote: %w", err)
	}
	defer res.Body.Close()
	const maxBody = 64 << 10
	body, err := io.ReadAll(io.LimitReader(res.Body, maxBody+1))
	if err != nil {
		return domain.Quote{}, fmt.Errorf("read quote: %w", err)
	}
	if len(body) > maxBody {
		return domain.Quote{}, fmt.Errorf("upstream body exceeds %d bytes", maxBody)
	}
	if res.StatusCode != http.StatusOK {
		return domain.Quote{}, fmt.Errorf("upstream HTTP status: %d", res.StatusCode)
	}
	var wire struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		return domain.Quote{}, fmt.Errorf("decode quote: %w", err)
	}
	if wire.Symbol != symbol {
		return domain.Quote{}, fmt.Errorf("upstream returned wrong symbol")
	}
	return domain.NewQuote(e.Source, wire.Symbol, wire.Price, e.Now())
}
