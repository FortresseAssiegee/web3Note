package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

func main() {
	// P-256 demonstrates ECDSA; it is NOT Ethereum's secp256k1.
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	message := []byte("APPROVE/v1|chain=111|asset=demo|amount=100|nonce=7")
	digest := sha256.Sum256(message)
	sig, err := ecdsa.SignASN1(rand.Reader, priv, digest[:])
	if err != nil {
		panic(err)
	}
	fmt.Println("valid signature:", ecdsa.VerifyASN1(&priv.PublicKey, digest[:], sig))
	wrong := sha256.Sum256([]byte("APPROVE/v1|chain=111|asset=demo|amount=900|nonce=7"))
	fmt.Println("changed amount accepted:", ecdsa.VerifyASN1(&priv.PublicKey, wrong[:], sig))
	other, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	fmt.Println("wrong public key accepted:", ecdsa.VerifyASN1(&other.PublicKey, digest[:], sig))
}
