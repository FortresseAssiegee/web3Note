package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
)

type Bloom struct {
	bits []byte
	m    uint64
	k    uint32
}

func newBloom(m uint64, k uint32) *Bloom {
	if m == 0 || k == 0 {
		panic("m and k must be positive")
	}
	return &Bloom{bits: make([]byte, (m+7)/8), m: m, k: k}
}

// A domain-separated SHA-256 for each position: educational, not Ethereum Bloom.
func (b *Bloom) positions(data []byte) []uint64 {
	result := make([]uint64, b.k)
	msg := make([]byte, 4+len(data))
	copy(msg[4:], data)
	for i := uint32(0); i < b.k; i++ {
		binary.BigEndian.PutUint32(msg[:4], i)
		h := sha256.Sum256(msg)
		result[i] = binary.BigEndian.Uint64(h[:8]) % b.m
	}
	return result
}

func (b *Bloom) add(data []byte) {
	for _, p := range b.positions(data) {
		b.bits[p/8] |= byte(1 << (p % 8))
	}
}

func (b *Bloom) maybeContains(data []byte) bool {
	for _, p := range b.positions(data) {
		if b.bits[p/8]&byte(1<<(p%8)) == 0 {
			return false
		}
	}
	return true
}

func optimal(n int, target float64) (uint64, uint32) {
	if n <= 0 || target <= 0 || target >= 1 {
		panic("invalid parameters")
	}
	m := uint64(math.Ceil(-float64(n) * math.Log(target) / (math.Ln2 * math.Ln2)))
	k := uint32(math.Round(float64(m) / float64(n) * math.Ln2))
	if k == 0 {
		k = 1
	}
	return m, k
}

func main() {
	m, k := optimal(1000, 0.01)
	fmt.Printf("n=1000 target=1%% -> m=%d bits, k=%d\n", m, k)
	approx := math.Pow(1-math.Exp(-float64(k)*1000/float64(m)), float64(k))
	fmt.Printf("approx_false_positive=%.4f\n", approx)
	b := newBloom(m, k)
	for _, name := range []string{"alice", "bob", "carol"} {
		b.add([]byte(name))
	}
	fmt.Println("alice:", b.maybeContains([]byte("alice")))
	fmt.Println("bob:", b.maybeContains([]byte("bob")))
	fmt.Println("not-inserted:", b.maybeContains([]byte("not-inserted")))
	// One bit deliberately demonstrates a false positive without randomness.
	tiny := newBloom(1, 1)
	tiny.add([]byte("alice"))
	fmt.Println("one-bit bloom says mallory may exist:", tiny.maybeContains([]byte("mallory")))
}
