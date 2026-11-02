package main

import (
	"bytes"
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

func main() {
	curve := ecdh.X25519()
	alice, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	bob, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	// This example assumes the peer public keys have already been authenticated.
	aliceSecret, err := alice.ECDH(bob.PublicKey())
	if err != nil {
		panic(err)
	}
	bobSecret, err := bob.ECDH(alice.PublicKey())
	if err != nil {
		panic(err)
	}
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		panic(err)
	}
	transcript := append([]byte("CHANNEL/v1|"), alice.PublicKey().Bytes()...)
	transcript = append(transcript, bob.PublicKey().Bytes()...)
	ctx := sha256.Sum256(transcript)
	info := fmt.Sprintf("Alice-to-Bob/aes256gcm/%x", ctx)
	ka, err := hkdf.Key(sha256.New, aliceSecret, salt, info, 32)
	if err != nil {
		panic(err)
	}
	kb, err := hkdf.Key(sha256.New, bobSecret, salt, info, 32)
	if err != nil {
		panic(err)
	}
	reverseInfo := fmt.Sprintf("Bob-to-Alice/aes256gcm/%x", ctx)
	reverse, err := hkdf.Key(sha256.New, aliceSecret, salt, reverseInfo, 32)
	if err != nil {
		panic(err)
	}
	fmt.Println("shared secret agrees:", bytes.Equal(aliceSecret, bobSecret))
	fmt.Println("derived keys agree:", bytes.Equal(ka, kb))
	fmt.Println("directions separated:", !bytes.Equal(ka, reverse))
}
