package main

import (
	"errors"
	"fmt"
	"math/big"
)

// Constant-product quote with a 0.3% input fee, matching the V2-style formula.
// This pure arithmetic example does not transfer assets or read a live pool.
func quote(amountIn, reserveIn, reserveOut *big.Int) (*big.Int, error) {
	if amountIn == nil || reserveIn == nil || reserveOut == nil || amountIn.Sign() <= 0 || reserveIn.Sign() <= 0 || reserveOut.Sign() <= 0 {
		return nil, errors.New("amount and reserves must be positive")
	}
	adjusted := new(big.Int).Mul(amountIn, big.NewInt(997))
	numerator := new(big.Int).Mul(adjusted, reserveOut)
	denominator := new(big.Int).Mul(reserveIn, big.NewInt(1000))
	denominator.Add(denominator, adjusted)
	return new(big.Int).Quo(numerator, denominator), nil
}

func mustQuote(a, x, y int64) *big.Int {
	q, err := quote(big.NewInt(a), big.NewInt(x), big.NewInt(y))
	if err != nil {
		panic(err)
	}
	return q
}

func check(name string, ok bool) {
	fmt.Printf("%s: %v\n", name, ok)
	if !ok {
		panic(name)
	}
}

func main() {
	// All values are integer base units; token decimals are a display convention.
	x, y := int64(1000000), int64(1000000)
	input, reserveIn, reserveOut := big.NewInt(10000), big.NewInt(x), big.NewInt(y)
	q, err := quote(input, reserveIn, reserveOut)
	if err != nil {
		panic(err)
	}
	fmt.Println("quoted output:", q)
	minOut := new(big.Int).Quo(new(big.Int).Mul(q, big.NewInt(99)), big.NewInt(100))
	fmt.Println("minimum output at 1% tolerance:", minOut)
	// An earlier trade changes the reserves before our trade executes.
	firstOut := mustQuote(100000, x, y).Int64()
	actual := mustQuote(10000, x+100000, y-firstOut)
	fmt.Println("output after earlier trade:", actual)
	check("minimum output rejects changed price", actual.Cmp(minOut) < 0)
	check("zero input rejected", func() bool { _, e := quote(big.NewInt(0), big.NewInt(x), big.NewInt(y)); return e != nil }())
	check("reserves remain caller-owned", input.Cmp(big.NewInt(10000)) == 0 && reserveIn.Cmp(big.NewInt(x)) == 0 && reserveOut.Cmp(big.NewInt(y)) == 0)
	kBefore := new(big.Int).Mul(big.NewInt(x), big.NewInt(y))
	kAfter := new(big.Int).Mul(big.NewInt(x+10000), new(big.Int).Sub(big.NewInt(y), q))
	check("fee-inclusive reserve product does not decrease", kAfter.Cmp(kBefore) >= 0)
}
