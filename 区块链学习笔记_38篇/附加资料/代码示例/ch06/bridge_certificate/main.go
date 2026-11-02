package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
)

// This is a committee-attestation model, not an Ethereum light client.
type Message struct {
	SourceChain, DestinationChain   uint64
	SourceBridge, DestinationBridge [32]byte
	CommitteeEpoch                  uint64
	FinalizedBlock, TxHash          [32]byte
	LogIndex                        uint32
	Recipient, Asset                [32]byte
	Amount                          uint64
}

func append64(dst []byte, n uint64) []byte {
	return binary.BigEndian.AppendUint64(dst, n)
}

func (m Message) SigningBytes() []byte {
	b := []byte("CryptoNotes/BridgeAttestation/v1\x00")
	b = append64(b, m.SourceChain)
	b = append64(b, m.DestinationChain)
	b = append(b, m.SourceBridge[:]...)
	b = append(b, m.DestinationBridge[:]...)
	b = append64(b, m.CommitteeEpoch)
	b = append(b, m.FinalizedBlock[:]...)
	b = append(b, m.TxHash[:]...)
	b = binary.BigEndian.AppendUint32(b, m.LogIndex)
	b = append(b, m.Recipient[:]...)
	b = append(b, m.Asset[:]...)
	return append64(b, m.Amount)
}

// The replay key identifies the source event, independently of its signature.
func (m Message) EventID() [32]byte {
	b := []byte("CryptoNotes/SourceEvent/v1\x00")
	b = append64(b, m.SourceChain)
	b = append(b, m.SourceBridge[:]...)
	b = append(b, m.TxHash[:]...)
	b = binary.BigEndian.AppendUint32(b, m.LogIndex)
	return sha256.Sum256(b)
}

type Witness struct {
	Member    int
	Signature []byte
}

type Receiver struct {
	mu       sync.Mutex
	Members  []ed25519.PublicKey // Fixed, trusted roster for this teaching epoch.
	Expected Message             // Only the domain fields and asset are used below.
	Used     map[[32]byte]bool
	Balances map[[32]byte]uint64
}

func (r *Receiver) Accept(m Message, witnesses []Witness) error {
	e := r.Expected
	if m.SourceChain != e.SourceChain || m.DestinationChain != e.DestinationChain ||
		m.SourceBridge != e.SourceBridge || m.DestinationBridge != e.DestinationBridge ||
		m.CommitteeEpoch != e.CommitteeEpoch || m.Asset != e.Asset {
		return errors.New("wrong domain, epoch, or asset")
	}
	if len(r.Members) != 4 || m.Amount == 0 || m.Recipient == ([32]byte{}) {
		return errors.New("invalid configuration or payload")
	}
	seen := make(map[int]bool)
	for _, w := range witnesses {
		if w.Member < 0 || w.Member >= len(r.Members) || seen[w.Member] {
			return errors.New("unknown or duplicate signer")
		}
		if !ed25519.Verify(r.Members[w.Member], m.SigningBytes(), w.Signature) {
			return errors.New("invalid signature")
		}
		seen[w.Member] = true
	}
	if len(seen) < 3 {
		return errors.New("need three distinct committee members")
	}

	// Atomic replay check and balance update within this one process.
	r.mu.Lock()
	defer r.mu.Unlock()
	id := m.EventID()
	if r.Used[id] {
		return errors.New("source event already consumed")
	}
	old := r.Balances[m.Recipient]
	if m.Amount > math.MaxUint64-old {
		return errors.New("balance overflow")
	}
	r.Balances[m.Recipient] = old + m.Amount
	r.Used[id] = true
	return nil
}

func check(name string, ok bool) {
	fmt.Printf("%s: %v\n", name, ok)
	if !ok {
		panic(name)
	}
}

func main() {
	pubs := make([]ed25519.PublicKey, 4)
	keys := make([]ed25519.PrivateKey, 4)
	for i := range keys {
		pub, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			panic(err)
		}
		pubs[i], keys[i] = pub, key
	}
	m := Message{SourceChain: 101, DestinationChain: 202, CommitteeEpoch: 7, Amount: 100}
	m.SourceBridge = sha256.Sum256([]byte("bridge-A"))
	m.DestinationBridge = sha256.Sum256([]byte("bridge-B"))
	m.FinalizedBlock = sha256.Sum256([]byte("claimed finalized block"))
	m.TxHash = sha256.Sum256([]byte("source transaction"))
	m.Recipient = sha256.Sum256([]byte("Bob"))
	m.Asset = sha256.Sum256([]byte("asset-A"))
	w := make([]Witness, 3)
	for i := range w {
		w[i] = Witness{Member: i, Signature: ed25519.Sign(keys[i], m.SigningBytes())}
	}
	fresh := func() *Receiver {
		return &Receiver{Members: pubs, Expected: m, Used: make(map[[32]byte]bool), Balances: make(map[[32]byte]uint64)}
	}
	check("two signatures rejected", fresh().Accept(m, w[:2]) != nil)
	check("duplicate signer rejected", fresh().Accept(m, []Witness{w[0], w[0], w[1]}) != nil)
	tampered := m
	tampered.Amount++
	check("changed amount rejected", fresh().Accept(tampered, w) != nil)
	wrongChain := m
	wrongChain.DestinationChain++
	check("other destination rejected", fresh().Accept(wrongChain, w) != nil)
	r := fresh()
	check("first event accepted", r.Accept(m, w) == nil)
	check("replay rejected", r.Accept(m, w) != nil)
	check("balance is 100", r.Balances[m.Recipient] == 100)
	over := fresh()
	over.Balances[m.Recipient] = math.MaxUint64
	check("overflow does not consume event", over.Accept(m, w) != nil && !over.Used[m.EventID()])
}
