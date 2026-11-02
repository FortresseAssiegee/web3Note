package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"example.com/quote-notes/internal/domain"
	"example.com/quote-notes/internal/service"
)

type API struct {
	service *service.Service
	timeout time.Duration
	ready   atomic.Bool
}

func New(s *service.Service, timeout time.Duration) *API {
	return &API{service: s, timeout: timeout}
}

func (a *API) SetReady(ready bool) { a.ready.Store(ready) }

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrSymbol):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unsupported_symbol"})
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "request_timeout"})
	default:
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "quote_unavailable"})
	}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if !a.ready.Load() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /v1/stats", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, a.service.Snapshot())
	})
	mux.HandleFunc("GET /v1/quotes/{symbol}", func(w http.ResponseWriter, r *http.Request) {
		if !a.ready.Load() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not_ready"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
		defer cancel()
		q, source, err := a.service.Latest(ctx, r.PathValue("symbol"))
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, struct {
			Quote  domain.Quote `json:"quote"`
			Source string       `json:"source"`
		}{q, source})
	})
	mux.HandleFunc("GET /v1/history/{symbol}", func(w http.ResponseWriter, r *http.Request) {
		if !a.ready.Load() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "not_ready"})
			return
		}
		limit := 20
		if raw := r.URL.Query().Get("limit"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > 100 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_limit"})
				return
			}
			limit = n
		}
		ctx, cancel := context.WithTimeout(r.Context(), a.timeout)
		defer cancel()
		rows, err := a.service.History(ctx, r.PathValue("symbol"), limit)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, rows)
	})
	return mux
}
