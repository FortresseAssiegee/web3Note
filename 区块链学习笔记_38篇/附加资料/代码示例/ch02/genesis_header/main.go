package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
)

func reverse(b []byte) []byte {
	out := append([]byte(nil), b...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// This helper accepts only positive, nonzero compact targets.
// It is a teaching decoder, not a full Bitcoin consensus validator.
func compactTarget(bits uint32) (*big.Int, error) {
	exponent := uint(bits >> 24)
	coefficient := bits & 0x007fffff
	if bits&0x00800000 != 0 || coefficient == 0 {
		return nil, fmt.Errorf("negative or zero target")
	}
	target := new(big.Int).SetUint64(uint64(coefficient))
	if exponent <= 3 {
		target.Rsh(target, 8*(3-exponent))
	} else {
		target.Lsh(target, 8*(exponent-3))
	}
	if target.Sign() <= 0 || target.BitLen() > 256 {
		return nil, fmt.Errorf("target out of range")
	}
	return target, nil
}

func main() {
	const merkleDisplay = "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	const expectedDisplay = "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"
	const bits uint32 = 0x1d00ffff

	merkle, err := hex.DecodeString(merkleDisplay)
	if err != nil {
		panic(err)
	}
	var header [80]byte
	binary.LittleEndian.PutUint32(header[0:4], 1)
	// header[4:36] remains all zero: no previous block.
	copy(header[36:68], reverse(merkle))
	binary.LittleEndian.PutUint32(header[68:72], 1231006505)
	binary.LittleEndian.PutUint32(header[72:76], bits)
	binary.LittleEndian.PutUint32(header[76:80], 2083236893)

	first := sha256.Sum256(header[:])
	second := sha256.Sum256(first[:])
	display := hex.EncodeToString(reverse(second[:]))
	target, err := compactTarget(bits)
	if err != nil {
		panic(err)
	}
	// big.Int.SetBytes reads big-endian, hence reverse the raw hash bytes.
	hashNumber := new(big.Int).SetBytes(reverse(second[:]))
	fmt.Println("header bytes:", len(header))
	fmt.Println("block hash:", display)
	fmt.Println("matches Bitcoin genesis:", display == expectedDisplay)
	fmt.Printf("target: %064x\n", target)
	fmt.Println("PoW hash <= target:", hashNumber.Cmp(target) <= 0)
}
