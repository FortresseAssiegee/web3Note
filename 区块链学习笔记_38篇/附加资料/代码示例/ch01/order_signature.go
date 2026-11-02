package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

// Order is an off-chain teaching message, not an Ethereum transaction.
type Order struct {
	ChainID   uint64
	Contract  [20]byte
	Recipient [20]byte
	Amount    uint64 // Smallest units: 25_000_000 means 25 units at 6 decimals.
	Nonce     uint64
	Expires   uint64 // Unix seconds; an order is valid at exactly Expires.
	InvoiceID [32]byte
}

func appendUint64(dst []byte, value uint64) []byte {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	return append(dst, encoded[:]...)
}

// Fixed field lengths and order make this teaching encoding unambiguous.
// It is NOT the EIP-712 encoding and is not intended as a new standard.
func encodeOrder(order Order) []byte {
	message := []byte("web3-notes/order/v1\x00")
	message = appendUint64(message, order.ChainID)
	message = append(message, order.Contract[:]...)
	message = append(message, order.Recipient[:]...)
	message = appendUint64(message, order.Amount)
	message = appendUint64(message, order.Nonce)
	message = appendUint64(message, order.Expires)
	return append(message, order.InvoiceID[:]...)
}

// This verifier is scoped to ONE buyer and used by ONE goroutine.
type Verifier struct {
	Buyer     ed25519.PublicKey
	ChainID   uint64
	Contract  [20]byte
	Recipient [20]byte
	Amount    uint64
	InvoiceID [32]byte
	Used      map[uint64]bool
}

func (v *Verifier) Accept(order Order, signature []byte, now uint64) error {
	if len(v.Buyer) != ed25519.PublicKeySize {
		return errors.New("invalid buyer public key")
	}
	if order.ChainID != v.ChainID || order.Contract != v.Contract {
		return errors.New("wrong application domain")
	}
	if order.Recipient != v.Recipient || order.Amount == 0 ||
		order.Amount != v.Amount || order.InvoiceID != v.InvoiceID {
		return errors.New("invalid payment terms")
	}
	if now > order.Expires {
		return errors.New("expired order")
	}
	if !ed25519.Verify(v.Buyer, encodeOrder(order), signature) {
		return errors.New("invalid signature")
	}
	if v.Used[order.Nonce] {
		return errors.New("nonce already used")
	}
	if v.Used == nil {
		v.Used = make(map[uint64]bool)
	}
	v.Used[order.Nonce] = true
	return nil
}

func main() {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	var contract, recipient [20]byte
	contract[19] = 1
	recipient[19] = 2
	const now uint64 = 1_800_000_000 // Fixed teaching clock, not the real clock.
	order := Order{
		ChainID: 31337, Contract: contract, Recipient: recipient,
		Amount: 25_000_000, Nonce: 7, Expires: now + 300,
		InvoiceID: sha256.Sum256([]byte("invoice-2026-001")),
	}
	signature := ed25519.Sign(privateKey, encodeOrder(order))
	verifier := Verifier{
		Buyer: publicKey, ChainID: 31337,
		Contract: contract, Recipient: recipient,
		Amount: 25_000_000, InvoiceID: order.InvoiceID,
		Used: make(map[uint64]bool),
	}
	fmt.Println("first accepted:", verifier.Accept(order, signature, now) == nil)
	fmt.Println("replay accepted:", verifier.Accept(order, signature, now) == nil)
	changed := order
	changed.Amount = 250_000_000
	fmt.Println("tampered signature valid:",
		ed25519.Verify(publicKey, encodeOrder(changed), signature))
	wrongDomain := order
	wrongDomain.ChainID = 1
	fmt.Println("other chain accepted:",
		verifier.Accept(wrongDomain, signature, now) == nil)
	expired := order
	expired.Nonce = 8
	expired.Expires = now - 1
	expiredSignature := ed25519.Sign(privateKey, encodeOrder(expired))
	fmt.Println("expired accepted:",
		verifier.Accept(expired, expiredSignature, now) == nil)
}
