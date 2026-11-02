package main

import (
	"fmt"
	"math/big"
)

var max256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

func number(s string) *big.Int {
	x, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("bad integer")
	}
	return x
}

func isU256(x *big.Int) bool { return x != nil && x.Sign() >= 0 && x.Cmp(max256) <= 0 }

type State struct {
	Balances map[string]*big.Int
	Nonces   map[string]uint64
	Burned   *big.Int
}

func (s State) clone() State {
	n := State{map[string]*big.Int{}, map[string]uint64{}, new(big.Int).Set(s.Burned)}
	for k, v := range s.Balances {
		n.Balances[k] = new(big.Int).Set(v)
	}
	for k, v := range s.Nonces {
		n.Nonces[k] = v
	}
	return n
}

func (s State) balance(who string) *big.Int {
	if v, ok := s.Balances[who]; ok {
		return new(big.Int).Set(v)
	}
	return new(big.Int)
}

func (s State) credit(who string, amount *big.Int) error {
	v := new(big.Int).Add(s.balance(who), amount)
	if !isU256(v) {
		return fmt.Errorf("recipient balance overflow")
	}
	s.Balances[who] = v
	return nil
}

type Tx struct {
	From, To        string
	Nonce, GasLimit uint64
	Value, FeeCap   *big.Int
	PriorityFeeCap  *big.Int
}

type Receipt struct {
	Success                   bool
	GasUsed                   uint64
	EffectivePrice, Fee, Burn *big.Int
	Tip                       *big.Int
}

// Teaching model for type-2 fee settlement, after authorization was checked.
// gasUsed/success are supplied execution outcomes, NOT computed by an EVM.
func apply(s State, tx Tx, baseFee *big.Int, gasUsed uint64, success bool) (State, Receipt, error) {
	bad := func(message string) (State, Receipt, error) {
		return s, Receipt{}, fmt.Errorf("%s", message)
	}
	if tx.From == "" || tx.To == "" {
		return bad("empty teaching account")
	}
	if !isU256(tx.Value) || !isU256(tx.FeeCap) || !isU256(tx.PriorityFeeCap) || !isU256(baseFee) {
		return bad("non-U256 input")
	}
	if tx.Nonce != s.Nonces[tx.From] || tx.Nonce == ^uint64(0) {
		return bad("wrong or exhausted nonce")
	}
	if tx.GasLimit < 21000 || gasUsed < 21000 || gasUsed > tx.GasLimit {
		return bad("invalid gas bounds in this demo")
	}
	if tx.PriorityFeeCap.Cmp(tx.FeeCap) > 0 || tx.FeeCap.Cmp(baseFee) < 0 {
		return bad("invalid fee caps")
	}
	upfront := new(big.Int).Mul(new(big.Int).SetUint64(tx.GasLimit), tx.FeeCap)
	upfront.Add(upfront, tx.Value)
	if !isU256(upfront) {
		return bad("upfront requirement exceeds U256")
	}
	if s.balance(tx.From).Cmp(upfront) < 0 {
		return bad("insufficient maximum upfront balance")
	}
	tipPerGas := new(big.Int).Sub(tx.FeeCap, baseFee)
	if tipPerGas.Cmp(tx.PriorityFeeCap) > 0 {
		tipPerGas.Set(tx.PriorityFeeCap)
	}
	price := new(big.Int).Add(baseFee, tipPerGas)
	gas := new(big.Int).SetUint64(gasUsed)
	fee := new(big.Int).Mul(gas, price)
	burn := new(big.Int).Mul(gas, baseFee)
	tip := new(big.Int).Mul(gas, tipPerGas)

	n := s.clone()
	n.Nonces[tx.From]++
	n.Balances[tx.From] = new(big.Int).Sub(n.balance(tx.From), fee)
	if success {
		n.Balances[tx.From].Sub(n.Balances[tx.From], tx.Value)
		if err := n.credit(tx.To, tx.Value); err != nil {
			return bad(err.Error())
		}
	}
	if err := n.credit("fee-recipient", tip); err != nil {
		return bad(err.Error())
	}
	n.Burned.Add(n.Burned, burn)
	return n, Receipt{success, gasUsed, price, fee, burn, tip}, nil
}

func main() {
	s := State{
		Balances: map[string]*big.Int{"alice": number("1000000000000000000")},
		Nonces:   map[string]uint64{}, Burned: new(big.Int),
	}
	tx := Tx{"alice", "bob", 0, 50000, number("100000000000000000"), number("35000000000"), number("3000000000")}
	base := number("20000000000")
	var r Receipt
	var err error
	s, r, err = apply(s, tx, base, 30000, false)
	if err != nil {
		panic(err)
	}
	fmt.Printf("revert: status=%t nonce=%d bob=%s fee=%s\n", r.Success, s.Nonces["alice"], s.balance("bob"), r.Fee)
	tx.Nonce = 1
	s, r, err = apply(s, tx, base, 21000, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("success: status=%t nonce=%d bob=%s fee=%s\n", r.Success, s.Nonces["alice"], s.balance("bob"), r.Fee)
	before := s.balance("alice")
	_, _, err = apply(s, tx, base, 21000, true)
	fmt.Println("replay rejected:", err != nil, "state unchanged:", before.Cmp(s.balance("alice")) == 0)
	tx.Nonce = 2
	tx.FeeCap = new(big.Int).Set(max256)
	_, _, err = apply(s, tx, base, 21000, true)
	fmt.Println("overflow rejected:", err)
	fmt.Printf("alice=%s bob=%s tips=%s burned=%s\n", s.balance("alice"), s.balance("bob"), s.balance("fee-recipient"), s.Burned)
}
