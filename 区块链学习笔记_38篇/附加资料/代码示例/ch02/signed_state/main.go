package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

type Address [ed25519.PublicKeySize]byte

type Account struct {
	Balance uint64
	Nonce   uint64
}

type Transaction struct {
	ChainID uint64
	From    Address
	To      Address
	Amount  uint64
	Nonce   uint64
}

func (t Transaction) message() []byte {
	var b bytes.Buffer
	b.WriteString("DEMO-LEDGER-TX-v1\x00")
	var n [8]byte
	for _, v := range []uint64{t.ChainID, t.Amount, t.Nonce} {
		binary.BigEndian.PutUint64(n[:], v)
		b.Write(n[:])
	}
	b.Write(t.From[:])
	b.Write(t.To[:])
	return b.Bytes()
}

type Ledger struct {
	ChainID  uint64
	Accounts map[Address]Account
}

func (l *Ledger) apply(t Transaction, signature []byte) error {
	if t.ChainID != l.ChainID {
		return fmt.Errorf("wrong chain")
	}
	if t.From == t.To || t.Amount == 0 {
		return fmt.Errorf("self-transfer or zero amount is excluded in this demo")
	}
	if !ed25519.Verify(ed25519.PublicKey(t.From[:]), t.message(), signature) {
		return fmt.Errorf("invalid signature")
	}
	from, exists := l.Accounts[t.From]
	if !exists {
		return fmt.Errorf("unknown sender")
	}
	to := l.Accounts[t.To]
	if t.Nonce != from.Nonce {
		return fmt.Errorf("wrong nonce: want %d, got %d", from.Nonce, t.Nonce)
	}
	if from.Balance < t.Amount {
		return fmt.Errorf("insufficient balance")
	}
	max := ^uint64(0)
	if to.Balance > max-t.Amount || from.Nonce == max {
		return fmt.Errorf("integer overflow")
	}
	from.Balance -= t.Amount
	from.Nonce++
	to.Balance += t.Amount
	// Only mutate state after every check succeeds.
	l.Accounts[t.From], l.Accounts[t.To] = from, to
	return nil
}

func keyPair() (Address, ed25519.PrivateKey) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	var address Address
	copy(address[:], pub)
	return address, priv
}

func main() {
	alice, aliceKey := keyPair()
	bob, _ := keyPair()
	ledger := Ledger{ChainID: 99, Accounts: map[Address]Account{
		alice: {Balance: 100}, bob: {Balance: 0},
	}}
	tx := Transaction{ChainID: 99, From: alice, To: bob, Amount: 30, Nonce: 0}
	sig := ed25519.Sign(aliceKey, tx.message())
	changed := tx
	changed.Amount = 80
	fmt.Println("tampered transaction:", ledger.apply(changed, sig))
	fmt.Println("first execution:", ledger.apply(tx, sig))
	fmt.Println("replay:", ledger.apply(tx, sig))
	overspend := Transaction{ChainID: 99, From: alice, To: bob, Amount: 80, Nonce: 1}
	overSig := ed25519.Sign(aliceKey, overspend.message())
	fmt.Println("signed overspend:", ledger.apply(overspend, overSig))
	fmt.Printf("Alice: balance=%d nonce=%d\n", ledger.Accounts[alice].Balance, ledger.Accounts[alice].Nonce)
	fmt.Printf("Bob: balance=%d nonce=%d\n", ledger.Accounts[bob].Balance, ledger.Accounts[bob].Nonce)
}
