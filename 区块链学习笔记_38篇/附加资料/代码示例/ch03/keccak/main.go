package main

import (
	"crypto/sha3"
	"fmt"
	legacy "golang.org/x/crypto/sha3"
)

func main() {
	standard := sha3.Sum256(nil)
	h := legacy.NewLegacyKeccak256()
	_, _ = h.Write(nil)
	fmt.Printf("SHA3-256(empty): %x\n", standard)
	fmt.Printf("Keccak256(empty): %x\n", h.Sum(nil))
}
