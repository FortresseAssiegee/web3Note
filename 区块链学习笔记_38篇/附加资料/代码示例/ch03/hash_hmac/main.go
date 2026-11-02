package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

// Each field carries its byte length; this is an educational protocol encoding.
func encodeFields(fields ...string) []byte {
	var b bytes.Buffer
	for _, s := range fields {
		if uint64(len(s)) > uint64(^uint32(0)) {
			panic("field too large")
		}
		if err := binary.Write(&b, binary.BigEndian, uint32(len(s))); err != nil {
			panic(err)
		}
		b.WriteString(s)
	}
	return b.Bytes()
}

func mac(key, msg []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write(msg)
	return h.Sum(nil)
}

func main() {
	a := []byte("ab" + "c")
	b := []byte("a" + "bc")
	fmt.Println("ambiguous concatenation:", sha256.Sum256(a) == sha256.Sum256(b))
	a = encodeFields("TRANSFER/v1", "ab", "c")
	b = encodeFields("TRANSFER/v1", "a", "bc")
	fmt.Println("structured encoding differs:", sha256.Sum256(a) != sha256.Sum256(b))
	fmt.Printf("SHA256(abc): %x\n", sha256.Sum256([]byte("abc")))
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	tag := mac(key, a)
	fmt.Println("valid HMAC:", hmac.Equal(tag, mac(key, a)))
	fmt.Println("tampered HMAC:", hmac.Equal(tag, mac(key, b)))
}
