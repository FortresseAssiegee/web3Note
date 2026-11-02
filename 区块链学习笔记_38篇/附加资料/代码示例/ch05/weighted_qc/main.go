// Verifies a teaching quorum certificate, not a complete consensus protocol.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
)

type Validator struct {
	ID     string
	Public ed25519.PublicKey
	Weight uint64
}

type Vote struct {
	Chain     string
	Epoch     uint64
	Height    uint64
	Round     uint64
	Step      byte
	Block     [32]byte
	SetDigest [32]byte
}

type SignedVote struct {
	Signer string
	Sig    []byte
}

func appendText(b []byte, s string) []byte {
	b = binary.BigEndian.AppendUint64(b, uint64(len(s)))
	return append(b, s...)
}

func registry(validators []Validator) ([32]byte, uint64, map[string]Validator, error) {
	ordered := append([]Validator(nil), validators...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	b := []byte("CH05-VALIDATOR-SET/v1\x00")
	b = binary.BigEndian.AppendUint64(b, uint64(len(ordered)))
	byID := make(map[string]Validator)
	byKey := make(map[string]bool)
	var total uint64
	for _, v := range ordered {
		_, duplicate := byID[v.ID]
		if v.ID == "" || duplicate || len(v.Public) != ed25519.PublicKeySize ||
			v.Weight == 0 || byKey[string(v.Public)] {
			return [32]byte{}, 0, nil, fmt.Errorf("invalid validator registry")
		}
		if ^uint64(0)-total < v.Weight {
			return [32]byte{}, 0, nil, fmt.Errorf("total weight overflow")
		}
		total += v.Weight
		byID[v.ID], byKey[string(v.Public)] = v, true
		b = appendText(b, v.ID)
		b = append(b, v.Public...)
		b = binary.BigEndian.AppendUint64(b, v.Weight)
	}
	if total == 0 {
		return [32]byte{}, 0, nil, fmt.Errorf("empty validator set")
	}
	return sha256.Sum256(b), total, byID, nil
}

func signBytes(v Vote, signer string) []byte {
	b := []byte("CH05-PRECOMMIT/v1\x00")
	b = appendText(b, v.Chain)
	b = binary.BigEndian.AppendUint64(b, v.Epoch)
	b = binary.BigEndian.AppendUint64(b, v.Height)
	b = binary.BigEndian.AppendUint64(b, v.Round)
	b = append(b, v.Step)
	b = append(b, v.Block[:]...)
	b = append(b, v.SetDigest[:]...)
	return appendText(b, signer)
}

func verifyQC(validators []Validator, expected Vote, signatures []SignedVote) (uint64, error) {
	digest, total, byID, err := registry(validators)
	if err != nil {
		return 0, err
	}
	if expected.Chain == "" || expected.Step != 2 || expected.SetDigest != digest {
		return 0, fmt.Errorf("wrong chain, step, or validator set")
	}
	seen := make(map[string]bool)
	var signed uint64
	for _, s := range signatures {
		v, ok := byID[s.Signer]
		if !ok || seen[s.Signer] {
			return 0, fmt.Errorf("unknown or duplicate signer")
		}
		seen[s.Signer] = true
		if !ed25519.Verify(v.Public, signBytes(expected, s.Signer), s.Sig) {
			return 0, fmt.Errorf("bad signature")
		}
		// Unique signer weights form a subset of the already overflow-checked total.
		signed += v.Weight
	}
	// floor(2*total/3)+1, without an overflowing multiplication.
	minimum := total - (total-1)/3
	if signed < minimum {
		return signed, fmt.Errorf("weight %d below strict threshold %d", signed, minimum)
	}
	return signed, nil
}

func main() {
	weights := []uint64{40, 30, 20, 10}
	var validators []Validator
	private := make(map[string]ed25519.PrivateKey)
	for i, weight := range weights {
		id := string(rune('A' + i))
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			panic(err)
		}
		validators = append(validators, Validator{id, pub, weight})
		private[id] = priv
	}
	digest, _, _, err := registry(validators)
	if err != nil {
		panic(err)
	}
	vote := Vote{Chain: "classroom-chain", Epoch: 7, Height: 88, Round: 2,
		Step: 2, Block: sha256.Sum256([]byte("block-88")), SetDigest: digest}
	sign := func(ids ...string) []SignedVote {
		out := make([]SignedVote, 0, len(ids))
		for _, id := range ids {
			out = append(out, SignedVote{id, ed25519.Sign(private[id], signBytes(vote, id))})
		}
		return out
	}
	qc := sign("A", "B")
	w, err := verifyQC(validators, vote, qc)
	fmt.Printf("A+B: weight=%d accepted=%t\n", w, err == nil)
	_, err = verifyQC(validators, vote, sign("A", "C"))
	fmt.Println("A+C below threshold rejected:", err != nil)
	_, err = verifyQC(validators, vote, sign("A", "A"))
	fmt.Println("duplicate A rejected:", err != nil)
	changed := vote
	changed.Round++
	_, err = verifyQC(validators, changed, qc)
	fmt.Println("changed round rejected:", err != nil)
	changed = vote
	changed.Block = sha256.Sum256([]byte("other-block"))
	_, err = verifyQC(validators, changed, qc)
	fmt.Println("changed block rejected:", err != nil)
	changed = vote
	changed.Chain = "another-chain"
	_, err = verifyQC(validators, changed, qc)
	fmt.Println("other chain rejected:", err != nil)
	alteredSet := append([]Validator(nil), validators...)
	alteredSet[0].Weight++
	_, err = verifyQC(alteredSet, vote, qc)
	fmt.Println("changed registry rejected:", err != nil)
}
