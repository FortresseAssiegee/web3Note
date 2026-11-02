package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/kzg4844"
)

func check(name string, ok bool) {
	fmt.Printf("%s: %v\n", name, ok)
	if !ok {
		panic(name)
	}
}

func main() {
	// Use the Go backend and the library's Ethereum trusted setup.
	if err := kzg4844.UseCKZG(false); err != nil {
		panic(err)
	}
	var blob kzg4844.Blob
	// Each 32-byte word must encode a canonical field element.
	// Here the first two evaluations are 1 and 2; the rest are zero.
	blob[31], blob[63] = 1, 2
	commitment, err := kzg4844.BlobToCommitment(&blob)
	if err != nil {
		panic(err)
	}
	proof, err := kzg4844.ComputeBlobProof(&blob, commitment)
	if err != nil {
		panic(err)
	}
	check("blob proof accepted", kzg4844.VerifyBlobProof(&blob, commitment, proof) == nil)
	changed := blob
	changed[31] = 3
	check("modified blob rejected", kzg4844.VerifyBlobProof(&changed, commitment, proof) != nil)

	var point kzg4844.Point
	point[31] = 7
	opening, value, err := kzg4844.ComputeProof(&blob, point)
	if err != nil {
		panic(err)
	}
	check("point opening accepted", kzg4844.VerifyProof(commitment, point, value, opening) == nil)
	badValue := value
	badValue[31] ^= 1
	check("modified evaluation rejected", kzg4844.VerifyProof(commitment, point, badValue, opening) != nil)

	versionedHash := kzg4844.CalcBlobHashV1(sha256.New(), &commitment)
	manual := sha256.Sum256(commitment[:])
	manual[0] = 1
	check("versioned hash agrees", manual == versionedHash)
	fmt.Printf("blob bytes=%d commitment bytes=%d proof bytes=%d\n", len(blob), len(commitment), len(proof))
	fmt.Printf("versioned hash=%x\n", versionedHash)
}
