package main

import (
	"crypto/rand"
	"fmt"

	blst "github.com/supranational/blst/bindings/go"
)

var sigDST = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")
var popDST = []byte("BLS_POP_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

func main() {
	message := []byte("VOTE/v1|chain=111|round=9|block=demo")
	pks := make([]*blst.P1Affine, 0, 3)
	sigs := make([]*blst.P2Affine, 0, 3)
	seen := make(map[string]bool)
	for i := 0; i < 3; i++ {
		ikm := make([]byte, 32)
		if _, err := rand.Read(ikm); err != nil {
			panic(err)
		}
		sk := blst.KeyGen(ikm)
		if sk == nil {
			panic("KeyGen failed")
		}
		pk := new(blst.P1Affine).From(sk)
		if !pk.KeyValidate() {
			panic("invalid key")
		}
		encoded := pk.Compress()
		if seen[string(encoded)] {
			panic("duplicate signer")
		}
		seen[string(encoded)] = true
		pop := new(blst.P2Affine).Sign(sk, encoded, popDST)
		if pop == nil || !pop.Verify(true, pk, true, encoded, popDST) {
			panic("invalid proof of possession")
		}
		sig := new(blst.P2Affine).Sign(sk, message, sigDST)
		if sig == nil || !sig.Verify(true, pk, true, message, sigDST) {
			panic("invalid signature")
		}
		pks = append(pks, pk)
		sigs = append(sigs, sig)
	}
	agg := new(blst.P2Aggregate)
	if !agg.Aggregate(sigs, true) {
		panic("aggregation failed")
	}
	signature := agg.ToAffine()
	fmt.Println("public key bytes:", len(pks[0].Compress()))
	fmt.Println("aggregate signature bytes:", len(signature.Compress()))
	// Fast verification is used only AFTER all public keys passed PoP verification.
	fmt.Println("aggregate valid:", signature.FastAggregateVerify(true, pks, message, sigDST))
	fmt.Println("changed message:", signature.FastAggregateVerify(true, pks, []byte("tampered"), sigDST))
	fmt.Println("missing signer:", signature.FastAggregateVerify(true, pks[:2], message, sigDST))
}
