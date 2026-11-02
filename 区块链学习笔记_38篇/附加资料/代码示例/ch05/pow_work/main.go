// Local teaching model: SHA-256 is truncated to 16 bits for the work check.
// This is not a Bitcoin node and implements no Bitcoin difficulty schedule.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type Block struct {
	Previous [32]byte
	Payload  string
	Target   uint16
	Nonce    uint64
	Hash     [32]byte
}

func headerHash(b Block) [32]byte {
	buf := append([]byte("CH05-TOY-POW/v1\x00"), b.Previous[:]...)
	buf = binary.BigEndian.AppendUint16(buf, b.Target)
	buf = binary.BigEndian.AppendUint64(buf, b.Nonce)
	buf = binary.BigEndian.AppendUint64(buf, uint64(len(b.Payload)))
	buf = append(buf, b.Payload...)
	return sha256.Sum256(buf)
}

func work(target uint16) uint64 {
	return (uint64(1) << 16) / (uint64(target) + 1)
}

func mine(previous [32]byte, payload string, target uint16) (Block, error) {
	b := Block{Previous: previous, Payload: payload, Target: target}
	// Hard cap: this demo never starts an unbounded mining job.
	for nonce := uint64(0); nonce < 1_000_000; nonce++ {
		b.Nonce = nonce
		b.Hash = headerHash(b)
		if binary.BigEndian.Uint16(b.Hash[:2]) <= b.Target {
			return b, nil
		}
	}
	return Block{}, fmt.Errorf("demo search limit reached")
}

func makeBranch(prefix string, count int, target uint16) ([]Block, error) {
	var previous [32]byte // Shared synthetic starting point.
	branch := make([]Block, 0, count)
	for i := 0; i < count; i++ {
		b, err := mine(previous, fmt.Sprintf("%s-%d", prefix, i+1), target)
		if err != nil {
			return nil, err
		}
		branch = append(branch, b)
		previous = b.Hash
	}
	return branch, nil
}

func verifyBranch(branch []Block) (uint64, error) {
	var previous [32]byte
	var total uint64
	for i, b := range branch {
		if b.Previous != previous || b.Hash != headerHash(b) {
			return 0, fmt.Errorf("block %d: broken link or altered header", i+1)
		}
		if binary.BigEndian.Uint16(b.Hash[:2]) > b.Target {
			return 0, fmt.Errorf("block %d: insufficient work", i+1)
		}
		w := work(b.Target)
		if ^uint64(0)-total < w {
			return 0, fmt.Errorf("work overflow")
		}
		total += w
		previous = b.Hash
	}
	return total, nil
}

func main() {
	a, err := makeBranch("easy", 3, 32767)
	if err != nil {
		panic(err)
	}
	b, err := makeBranch("hard", 2, 4095)
	if err != nil {
		panic(err)
	}
	wa, err := verifyBranch(a)
	if err != nil {
		panic(err)
	}
	wb, err := verifyBranch(b)
	if err != nil {
		panic(err)
	}
	fmt.Printf("A: blocks=%d work=%d\n", len(a), wa)
	fmt.Printf("B: blocks=%d work=%d\n", len(b), wb)
	fmt.Println("choose B by work:", wb > wa)
	a[0].Payload = "tampered"
	_, err = verifyBranch(a)
	fmt.Println("tamper rejected:", err != nil)
}
