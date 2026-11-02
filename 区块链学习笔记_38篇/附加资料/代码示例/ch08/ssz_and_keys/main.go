package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type Root [32]byte

func hashPair(a, b Root) Root {
	var input [64]byte
	copy(input[:32], a[:])
	copy(input[32:], b[:])
	return sha256.Sum256(input[:])
}

// Fixed limit of 8 uint64 values: basic elements pack four per chunk.
func chunks(values []uint64) ([2]Root, error) {
	var out [2]Root
	if len(values) > 8 {
		return out, fmt.Errorf("limit 8 exceeded")
	}
	for i, v := range values {
		binary.LittleEndian.PutUint64(out[i/4][(i%4)*8:], v)
	}
	return out, nil
}

func vectorRoot(values []uint64) (Root, error) {
	if len(values) != 8 {
		return Root{}, fmt.Errorf("Vector[uint64,8] requires 8 values")
	}
	c, err := chunks(values)
	if err != nil {
		return Root{}, err
	}
	return hashPair(c[0], c[1]), nil
}

func listRoot(values []uint64) (Root, error) {
	c, err := chunks(values)
	if err != nil {
		return Root{}, err
	}
	var length Root
	binary.LittleEndian.PutUint64(length[:8], uint64(len(values)))
	return hashPair(hashPair(c[0], c[1]), length), nil
}

// RLP of a nonnegative integer, sufficient for a transaction/receipt index.
// This does not build an MPT and does not hash an Ethereum state key.
func rlpIndex(v uint64) []byte {
	if v == 0 {
		return []byte{0x80}
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], v)
	i := 0
	for b[i] == 0 {
		i++
	}
	data := b[i:]
	if len(data) == 1 && data[0] < 0x80 {
		return append([]byte(nil), data...)
	}
	out := []byte{0x80 + byte(len(data))}
	return append(out, data...)
}

func main() {
	v := []uint64{1, 2, 3, 4, 0, 0, 0, 0}
	vr, err := vectorRoot(v)
	if err != nil {
		panic(err)
	}
	lr4, err := listRoot(v[:4])
	if err != nil {
		panic(err)
	}
	lr8, err := listRoot(v)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Vector[uint64,8] root: %x\n", vr)
	fmt.Printf("List[uint64,8] length4 root: %x\n", lr4)
	fmt.Printf("List[uint64,8] length8 root: %x\n", lr8)
	fmt.Println("same padded data, different lengths:", lr4 != lr8)
	c, _ := chunks(v)
	fmt.Println("two-chunk vector reconstruction:", hashPair(c[0], c[1]) == vr)
	c[1][0] = 1
	fmt.Println("tampered chunk rejected:", hashPair(c[0], c[1]) != vr)
	_, err = listRoot(make([]uint64, 9))
	fmt.Println("over limit rejected:", err != nil)
	for _, i := range []uint64{0, 1, 127, 128, 256} {
		fmt.Printf("MPT index %d -> RLP %x\n", i, rlpIndex(i))
	}
}
