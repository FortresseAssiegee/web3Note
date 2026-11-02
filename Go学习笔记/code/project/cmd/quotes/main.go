package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/quote-notes/internal/config"
	"example.com/quote-notes/internal/exchange"
	"example.com/quote-notes/internal/httpapi"
	"example.com/quote-notes/internal/memory"
	"example.com/quote-notes/internal/service"
)

func demoUpstream() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v3/ticker/price" {
			http.NotFound(w, r)
			return
		}
		symbol := r.URL.Query().Get("symbol")
		price, ok := map[string]string{"BTCUSDT": "60000.10", "ETHUSDT": "3000.20"}[symbol]
		if !ok {
			http.Error(w, "unknown symbol", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		}{symbol, price})
	}))
}

func run(args []string) error {
	cfg, err := config.Parse(args)
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	baseURL := "https://data-api.binance.vision"
	source := "binance-spot"
	if cfg.Mode == "demo" {
		mock := demoUpstream()
		defer mock.Close()
		baseURL = mock.URL
		source = "demo-binance-spot"
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxConnsPerHost = 4
	transport.MaxIdleConnsPerHost = 4
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	e := &exchange.HTTP{Source: source, BaseURL: baseURL, Client: client, Now: time.Now}
	svc := service.New(e, memory.NewCache(time.Now), memory.NewHistory(1000), cfg.TTL)
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	for _, symbol := range symbols {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		_, err = svc.Refresh(ctx, symbol)
		cancel()
		if err != nil {
			return fmt.Errorf("startup warmup: %w", err)
		}
	}
	api := httpapi.New(svc, cfg.Timeout)
	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	server := &http.Server{Handler: api.Handler(), ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout: cfg.Timeout + time.Second, IdleTimeout: 30 * time.Second}
	defer server.Close()
	api.SetReady(true)
	collectorCtx, stopCollector := context.WithCancel(context.Background())
	collectorDone := make(chan struct{})
	go func() {
		defer close(collectorDone)
		service.RunCollector(collectorCtx, svc, symbols, cfg.Interval, cfg.Timeout,
			func(err error) { logger.Error("sample failed", "error", err) })
	}()
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	logger.Info("ready", "addr", listener.Addr().String(), "mode", cfg.Mode)
	var serveErr error
	serveReceived := false
	select {
	case <-signalCtx.Done():
	case serveErr = <-serveDone:
		serveReceived = true
	}
	api.SetReady(false)
	stopCollector()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	shutdownErr := server.Shutdown(shutdownCtx)
	cancel()
	if shutdownErr != nil {
		_ = server.Close()
	}
	<-collectorDone
	if !serveReceived {
		serveErr = <-serveDone
	}
	logger.Info("stopped", "stats", svc.Snapshot())
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return serveErr
	}
	return shutdownErr
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "quotes:", err)
		os.Exit(1)
	}
}
