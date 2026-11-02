package main

import (
	"crypto/sha256"
	"fmt"
)

type Hash [32]byte

// Step records the sibling hash and its position relative to the current node.
type Step struct {
	Sibling Hash
	Left    bool
}

func leaf(data []byte) Hash {
	b := append([]byte{0}, data...)
	return sha256.Sum256(b)
}

func branch(left, right Hash) Hash {
	b := make([]byte, 1, 65)
	b[0] = 1
	b = append(b, left[:]...)
	b = append(b, right[:]...)
	return sha256.Sum256(b)
}

// Split at the largest power of two strictly smaller than n; n must exceed 1.
func split(n int) int {
	k := 1
	for k <= (n-1)/2 {
		k *= 2
	}
	return k
}

func root(items [][]byte) Hash {
	switch len(items) {
	case 0:
		return sha256.Sum256(nil)
	case 1:
		return leaf(items[0])
	default:
		k := split(len(items))
		return branch(root(items[:k]), root(items[k:]))
	}
}

func prove(items [][]byte, index int) ([]Step, error) {
	if index < 0 || index >= len(items) {
		return nil, fmt.Errorf("index out of range")
	}
	if len(items) == 1 {
		return nil, nil
	}
	k := split(len(items))
	if index < k {
		p, err := prove(items[:k], index)
		return append(p, Step{root(items[k:]), false}), err
	}
	p, err := prove(items[k:], index-k)
	return append(p, Step{root(items[:k]), true}), err
}

// directions derives the expected proof shape from the trusted size and index.
func directions(size, index int) []bool {
	if size == 1 {
		return nil
	}
	k := split(size)
	if index < k {
		return append(directions(k, index), false)
	}
	return append(directions(size-k, index-k), true)
}

func verify(data []byte, size, index int, proof []Step, expected Hash) bool {
	if size <= 0 || index < 0 || index >= size {
		return false
	}
	dirs := directions(size, index)
	if len(dirs) != len(proof) {
		return false
	}
	h := leaf(data)
	for i, s := range proof {
		if s.Left != dirs[i] {
			return false
		}
		if s.Left {
			h = branch(s.Sibling, h)
		} else {
			h = branch(h, s.Sibling)
		}
	}
	return h == expected
}

func main() {
	items := [][]byte{[]byte("A:10"), []byte("B:20"), []byte("C:30"), []byte("D:40"), []byte("E:50")}
	r := root(items)
	p, err := prove(items, 2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("root=%x\nproof_steps=%d\n", r, len(p))
	fmt.Println("valid:", verify([]byte("C:30"), 5, 2, p, r))
	fmt.Println("tampered value:", verify([]byte("C:300"), 5, 2, p, r))
	fmt.Println("wrong index:", verify([]byte("C:30"), 5, 1, p, r))
	bad := append([]Step(nil), p...)
	bad[0].Sibling[0] ^= 1
	fmt.Println("tampered sibling:", verify([]byte("C:30"), 5, 2, bad, r))
}
