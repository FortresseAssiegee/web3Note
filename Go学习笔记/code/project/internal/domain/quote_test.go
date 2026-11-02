package domain

import (
	"testing"
	"time"
)

func TestPriceContract(t *testing.T) {
	for _, price := range []string{"0.01", "60000.1000", "0001.25", "99999999999999999999.123456789012345678"} {
		if err := ValidatePrice(price); err != nil {
			t.Fatalf("%q: %v", price, err)
		}
	}
	for _, price := range []string{"", "0", "-1", "NaN", "1e3", "1.", ".1", "1/2", "1 2", "100000000000000000000", "0.1234567890123456789"} {
		if ValidatePrice(price) == nil {
			t.Errorf("accepted invalid price %q", price)
		}
	}
}

func TestObservationIsNotAnExchangeEvent(t *testing.T) {
	when := time.Date(2026, 10, 22, 10, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	a, err := NewQuote("demo-binance-spot", "BTCUSDT", "60000.10", when)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewQuote("demo-binance-spot", "BTCUSDT", "60000.10", when)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.ID) != 32 || a.ID == b.ID {
		t.Fatal("observation IDs were not distinct")
	}
	if a.ObservedAt.Location() != time.UTC || !a.ObservedAt.Equal(when) {
		t.Fatal("timestamp changed meaning")
	}
	if _, err := NewQuote("demo", "btcUSDT", "1", when); err != ErrSymbol {
		t.Fatal("symbol contract lost")
	}
}
