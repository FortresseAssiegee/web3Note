package config

import (
	"errors"
	"flag"
	"io"
	"time"
)

type Config struct {
	Addr     string
	Mode     string
	Interval time.Duration
	TTL      time.Duration
	Timeout  time.Duration
}

func Parse(args []string) (Config, error) {
	var c Config
	f := flag.NewFlagSet("quotes", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.StringVar(&c.Addr, "addr", "127.0.0.1:8080", "HTTP listen address")
	f.StringVar(&c.Mode, "mode", "demo", "demo or public")
	f.DurationVar(&c.Interval, "interval", 2*time.Second, "sampling interval")
	f.DurationVar(&c.TTL, "ttl", 5*time.Second, "cache TTL")
	f.DurationVar(&c.Timeout, "timeout", 2*time.Second, "request timeout")
	if err := f.Parse(args); err != nil {
		return c, err
	}
	if f.NArg() != 0 {
		return c, errors.New("unexpected positional arguments")
	}
	if c.Mode != "demo" && c.Mode != "public" {
		return c, errors.New("mode must be demo or public")
	}
	if c.Interval <= 0 || c.TTL <= 0 || c.Timeout <= 0 {
		return c, errors.New("durations must be positive")
	}
	return c, nil
}
