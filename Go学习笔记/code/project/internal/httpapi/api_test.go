package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/quote-notes/internal/domain"
	"example.com/quote-notes/internal/memory"
	"example.com/quote-notes/internal/service"
)

type exchangeStub struct{}

func (exchangeStub) Fetch(ctx context.Context, symbol string) (domain.Quote, error) {
	return domain.Quote{ID: "demo", Exchange: "mock", Symbol: symbol, Price: "10.00", ObservedAt: time.Unix(1, 0).UTC()}, nil
}

func TestHTTPBoundaries(t *testing.T) {
	s := service.New(exchangeStub{}, memory.NewCache(time.Now), memory.NewHistory(10), time.Minute)
	a := New(s, time.Second)
	handler := a.Handler()
	r := httptest.NewRecorder()
	handler.ServeHTTP(r, httptest.NewRequest("GET", "/readyz", nil))
	if r.Code != 503 {
		t.Fatal("must not be ready before initialization")
	}
	a.SetReady(true)
	for _, tt := range []struct {
		method, path string
		status       int
		want         string
	}{
		{"GET", "/v1/quotes/BTCUSDT", 200, `"symbol":"BTCUSDT"`},
		{"GET", "/v1/quotes/BTCUSDT?symbol=ETHUSDT", 200, `"symbol":"BTCUSDT"`},
		{"GET", "/v1/quotes/NOPE", 404, `"unsupported_symbol"`},
		{"POST", "/v1/quotes/BTCUSDT", 405, ""},
		{"GET", "/v1/history/BTCUSDT?limit=0", 400, `"invalid_limit"`},
		{"GET", "/v1/history/BTCUSDT?limit=1", 200, `"price":"10.00"`},
	} {
		r := httptest.NewRecorder()
		handler.ServeHTTP(r, httptest.NewRequest(tt.method, tt.path, nil))
		if r.Code != tt.status || !strings.Contains(r.Body.String(), tt.want) {
			t.Errorf("%s %s: code=%d body=%s", tt.method, tt.path, r.Code, r.Body.String())
		}
	}
	a.SetReady(false)
	r = httptest.NewRecorder()
	handler.ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/v1/quotes/BTCUSDT", nil))
	if r.Code != 503 {
		t.Fatal("draining service still accepted quote requests")
	}
}
