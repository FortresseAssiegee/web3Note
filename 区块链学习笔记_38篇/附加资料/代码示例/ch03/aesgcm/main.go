package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

func main() {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	// Go 1.24+: nonce is generated and prepended by Seal.
	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		panic(err)
	}
	msg := []byte("invoice: Alice pays Bob 100 units")
	aad := []byte("INVOICE/v1|chain=111|contract=demo|record=7")
	sealed := aead.Seal(nil, nil, msg, aad)
	plain, err := aead.Open(nil, nil, sealed, aad)
	if err != nil {
		panic(err)
	}
	fmt.Println("recovered:", string(plain))
	fmt.Println("overhead:", len(sealed)-len(msg))
	bad := append([]byte(nil), sealed...)
	bad[len(bad)-1] ^= 1
	_, err = aead.Open(nil, nil, bad, aad)
	fmt.Println("tampered ciphertext rejected:", err != nil)
	_, err = aead.Open(nil, nil, sealed, []byte("wrong context"))
	fmt.Println("wrong AAD rejected:", err != nil)
	wrongKey := make([]byte, 32)
	if _, err := rand.Read(wrongKey); err != nil {
		panic(err)
	}
	wrongBlock, err := aes.NewCipher(wrongKey)
	if err != nil {
		panic(err)
	}
	wrongAEAD, err := cipher.NewGCMWithRandomNonce(wrongBlock)
	if err != nil {
		panic(err)
	}
	_, err = wrongAEAD.Open(nil, nil, sealed, aad)
	fmt.Println("wrong key rejected:", err != nil)
}
