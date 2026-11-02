package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// Tiny integers ONLY illustrate algebra. They provide no cryptographic security.
const p, q, g int64 = 23, 11, 2

func mod(x, n int64) int64 { return (x%n + n) % n }
func pow(a, n int64) int64 {
	r := int64(1)
	for n > 0 {
		if n&1 != 0 {
			r = r * a % p
		}
		a = a * a % p
		n >>= 1
	}
	return r
}
func challenge(y, r int64, message string) int64 {
	b := []byte("TOY-SCHNORR/v1")
	for _, v := range []int64{p, q, g, y, r} {
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(v))
		b = append(b, buf[:]...)
	}
	b = append(b, []byte(message)...)
	h := sha256.Sum256(b)
	e := int64(0)
	for _, v := range h {
		e = (e*256 + int64(v)) % q
	}
	return e
}
func verify(y, r, s int64, message string) bool {
	if y <= 1 || y >= p || r <= 1 || r >= p || s < 0 || s >= q {
		return false
	}
	if pow(y, q) != 1 || pow(r, q) != 1 {
		return false
	}
	e := challenge(y, r, message)
	return pow(g, s) == r*pow(y, e)%p
}
func main() {
	x, k := int64(3), int64(4) // Deliberately fixed only for this tiny toy.
	y, r := pow(g, x), pow(g, k)
	m1 := "pay Bob 10"
	e1 := challenge(y, r, m1)
	s1 := mod(k+e1*x, q)
	// A tiny q causes many challenge collisions; select a different challenge.
	m2 := ""
	e2 := e1
	for i := 0; e2 == e1; i++ {
		m2 = fmt.Sprintf("changed message %d", i)
		e2 = challenge(y, r, m2)
	}
	s2 := mod(k+e2*x, q) // Intentional nonce reuse to demonstrate the failure.
	den := mod(e1-e2, q)
	inv := int64(1)
	for den*inv%q != 1 {
		inv++
	}
	recovered := mod((s1-s2)*inv, q)
	fmt.Println("valid:", verify(y, r, s1, m1))
	fmt.Println("changed message:", verify(y, r, s1, m2))
	fmt.Println("invalid subgroup:", verify(5, r, s1, m1))
	fmt.Println("secret recovered after nonce reuse:", recovered)
}
