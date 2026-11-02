package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

func main() {
	priv, err := btcec.NewPrivateKey()
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256([]byte("DEMO/v1|pay=100|nonce=7"))
	var aux [32]byte
	if _, err := rand.Read(aux[:]); err != nil {
		panic(err)
	}
	sig, err := schnorr.Sign(priv, digest[:], schnorr.CustomNonce(aux))
	if err != nil {
		panic(err)
	}
	pkBytes := schnorr.SerializePubKey(priv.PubKey())
	pk, err := schnorr.ParsePubKey(pkBytes)
	if err != nil {
		panic(err)
	}
	fmt.Println("public key bytes:", len(pkBytes))
	fmt.Println("signature bytes:", len(sig.Serialize()))
	fmt.Println("valid:", sig.Verify(digest[:], pk))
	wrong := sha256.Sum256([]byte("DEMO/v1|pay=900|nonce=7"))
	fmt.Println("changed amount:", sig.Verify(wrong[:], pk))
}
