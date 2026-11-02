package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"
)

var ErrSymbol = errors.New("unsupported symbol")
var decimal = regexp.MustCompile(`^[0-9]+(?:\.[0-9]+)?$`)

type Quote struct {
	ID         string    `json:"id"`
	Exchange   string    `json:"exchange"`
	Symbol     string    `json:"symbol"`
	Price      string    `json:"price"`
	ObservedAt time.Time `json:"observed_at"`
}

func ValidateSymbol(symbol string) error {
	if symbol != "BTCUSDT" && symbol != "ETHUSDT" {
		return ErrSymbol
	}
	return nil
}

func ValidatePrice(price string) error {
	if len(price) == 0 || len(price) > 64 || !decimal.MatchString(price) {
		return errors.New("price must be a positive decimal string")
	}
	parts := strings.SplitN(price, ".", 2)
	if len(strings.TrimLeft(parts[0], "0")) > 20 || (len(parts) == 2 && len(parts[1]) > 18) {
		return errors.New("price exceeds DECIMAL(38,18) precision")
	}
	n, ok := new(big.Rat).SetString(price)
	if !ok || n.Sign() <= 0 {
		return errors.New("price must be positive")
	}
	return nil
}

func NewQuote(exchange, symbol, price string, observedAt time.Time) (Quote, error) {
	if exchange == "" {
		return Quote{}, errors.New("exchange name is required")
	}
	if err := ValidateSymbol(symbol); err != nil {
		return Quote{}, err
	}
	if err := ValidatePrice(price); err != nil {
		return Quote{}, err
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return Quote{}, err
	}
	return Quote{ID: hex.EncodeToString(id[:]), Exchange: exchange,
		Symbol: symbol, Price: price, ObservedAt: observedAt.UTC()}, nil
}
