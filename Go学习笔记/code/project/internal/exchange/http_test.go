package exchange

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPResponseContract(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		good   bool
	}{
		{"good", 200, `{"symbol":"BTCUSDT","price":"123.4500","extra":"allowed"}`, true},
		{"status", 503, `{}`, false},
		{"malformed", 200, `{`, false},
		{"wrong_symbol", 200, `{"symbol":"ETHUSDT","price":"1"}`, false},
		{"negative", 200, `{"symbol":"BTCUSDT","price":"-1"}`, false},
		{"two_values", 200, `{"symbol":"BTCUSDT","price":"1"}{}`, false},
		{"huge", 200, strings.Repeat("x", (64<<10)+1), false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v3/ticker/price" || r.URL.Query().Get("symbol") != "BTCUSDT" {
					t.Error("incorrect request")
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			e := &HTTP{Source: "demo", BaseURL: server.URL, Client: server.Client(), Now: func() time.Time { return time.Unix(100, 0) }}
			q, err := e.Fetch(context.Background(), "BTCUSDT")
			if (err == nil) != tt.good {
				t.Fatalf("good=%v err=%v", tt.good, err)
			}
			if tt.good && (q.Price != "123.4500" || q.Exchange != "demo") {
				t.Fatal("lost price or source")
			}
		})
	}
}

func TestHTTPCancellation(t *testing.T) {
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
	}))
	defer server.Close()
	e := &HTTP{Source: "demo", BaseURL: server.URL, Client: server.Client(), Now: time.Now}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := e.Fetch(ctx, "BTCUSDT"); done <- err }()
	select {
	case <-entered:
	case err := <-done:
		t.Fatalf("request failed early: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("handler not reached")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("request did not stop")
	}
}
