package main

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

type Account struct{ Balance, Nonce uint64 }
type State [2]Account
type Transfer struct {
	From, To      uint8
	Amount, Nonce uint64
}

func stateRoot(s State) [32]byte {
	b := []byte("CryptoNotes/State/v1\x00")
	for i, a := range s {
		b = append(b, byte(i))
		b = binary.BigEndian.AppendUint64(b, a.Balance)
		b = binary.BigEndian.AppendUint64(b, a.Nonce)
	}
	return sha256.Sum256(b)
}

func batchCommitment(batch []Transfer) [32]byte {
	b := []byte("CryptoNotes/Batch/v1\x00")
	b = binary.BigEndian.AppendUint64(b, uint64(len(batch)))
	for _, t := range batch {
		b = append(b, t.From, t.To)
		b = binary.BigEndian.AppendUint64(b, t.Amount)
		b = binary.BigEndian.AppendUint64(b, t.Nonce)
	}
	return sha256.Sum256(b)
}

// Authorization is assumed to have been checked before this teaching function.
// This is full re-execution; it is neither a fraud-proof VM nor a ZK proof.
func execute(s State, batch []Transfer) (State, error) {
	next := s
	for _, t := range batch {
		if t.From > 1 || t.To > 1 || t.From == t.To || t.Amount == 0 {
			return s, errors.New("bad transfer")
		}
		from, to := &next[t.From], &next[t.To]
		if from.Nonce != t.Nonce || from.Nonce == math.MaxUint64 || from.Balance < t.Amount || to.Balance > math.MaxUint64-t.Amount {
			return s, errors.New("nonce, balance, or overflow")
		}
		from.Balance -= t.Amount
		from.Nonce++
		to.Balance += t.Amount
	}
	return next, nil
}

func validate(old State, data []Transfer, available bool, commitment, claimedRoot [32]byte) error {
	if !available {
		return errors.New("data unavailable: cannot re-execute")
	}
	if batchCommitment(data) != commitment {
		return errors.New("batch commitment mismatch")
	}
	next, err := execute(old, data)
	if err != nil {
		return err
	}
	if stateRoot(next) != claimedRoot {
		return errors.New("incorrect state transition")
	}
	return nil
}

func check(name string, ok bool) {
	fmt.Printf("%s: %v\n", name, ok)
	if !ok {
		panic(name)
	}
}

func main() {
	old := State{{Balance: 100}, {Balance: 20}}
	batch := []Transfer{{From: 0, To: 1, Amount: 10, Nonce: 0}, {From: 0, To: 1, Amount: 5, Nonce: 1}}
	next, err := execute(old, batch)
	if err != nil {
		panic(err)
	}
	c, root := batchCommitment(batch), stateRoot(next)
	check("valid batch accepted", validate(old, batch, true, c, root) == nil)
	wrong := next
	wrong[1].Balance++
	check("committed data cannot justify wrong state", validate(old, batch, true, c, stateRoot(wrong)) != nil)
	check("commitment alone is not available data", validate(old, nil, false, c, root) != nil)
	reversed := []Transfer{batch[1], batch[0]}
	check("order changes commitment", batchCommitment(reversed) != c)
	_, err = execute(old, reversed)
	check("reordered nonces rejected", err != nil)
	fmt.Printf("final: Alice=%d nonce=%d Bob=%d\n", next[0].Balance, next[0].Nonce, next[1].Balance)
}
