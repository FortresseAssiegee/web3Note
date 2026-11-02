// 教学模型：使用 Ed25519 演示授权和 UTXO 状态转换，不是 Bitcoin 交易格式。
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const maxMoney uint64 = 21_000_000 * 100_000_000

type Outpoint struct {
	ID    [32]byte
	Index uint32
}
type Coin struct {
	Value uint64
	Owner [32]byte
}
type Tx struct {
	Inputs  []Outpoint
	Outputs []Coin
	Sigs    [][]byte
}
type View map[Outpoint]Coin

func put32(b *bytes.Buffer, n uint32) { _ = binary.Write(b, binary.LittleEndian, n) }
func put64(b *bytes.Buffer, n uint64) { _ = binary.Write(b, binary.LittleEndian, n) }

func body(tx Tx) []byte {
	var b bytes.Buffer
	b.WriteString("ch07/utxo/v1/test-network") // 明确演示协议及网络的域
	put32(&b, uint32(len(tx.Inputs)))
	for _, in := range tx.Inputs {
		b.Write(in.ID[:])
		put32(&b, in.Index)
	}
	put32(&b, uint32(len(tx.Outputs)))
	for _, out := range tx.Outputs {
		put64(&b, out.Value)
		b.Write(out.Owner[:])
	}
	return b.Bytes()
}

func digest(tx Tx, v View, index int) ([32]byte, error) {
	var b bytes.Buffer
	b.WriteString("ch07/sign-all-prevouts/v1")
	b.Write(body(tx))
	put32(&b, uint32(index))
	for _, in := range tx.Inputs {
		coin, ok := v[in]
		if !ok {
			return [32]byte{}, fmt.Errorf("missing or spent input")
		}
		put64(&b, coin.Value)
		b.Write(coin.Owner[:])
	}
	return sha256.Sum256(b.Bytes()), nil
}

func add(sum, value uint64) (uint64, error) {
	// 在加法之前检查，避免机器整数先溢出再比较。
	if value > maxMoney || sum > maxMoney-value {
		return 0, fmt.Errorf("money range exceeded")
	}
	return sum + value, nil
}

func apply(v View, tx Tx) (uint64, error) {
	if len(tx.Inputs) == 0 || len(tx.Inputs) > 1000 ||
		len(tx.Outputs) == 0 || len(tx.Outputs) > 1000 || len(tx.Sigs) != len(tx.Inputs) {
		return 0, fmt.Errorf("invalid vector sizes")
	}
	seen := make(map[Outpoint]bool)
	var inSum, outSum uint64
	for i, in := range tx.Inputs {
		if seen[in] {
			return 0, fmt.Errorf("duplicate input")
		}
		seen[in] = true
		coin, ok := v[in]
		if !ok {
			return 0, fmt.Errorf("missing or spent input")
		}
		var err error
		inSum, err = add(inSum, coin.Value)
		if err != nil {
			return 0, err
		}
		d, err := digest(tx, v, i)
		if err != nil || !ed25519.Verify(ed25519.PublicKey(coin.Owner[:]), d[:], tx.Sigs[i]) {
			return 0, fmt.Errorf("invalid signature at input %d", i)
		}
	}
	for _, out := range tx.Outputs {
		var err error
		outSum, err = add(outSum, out.Value)
		if err != nil {
			return 0, err
		}
	}
	if outSum > inSum {
		return 0, fmt.Errorf("outputs exceed inputs")
	}
	id := sha256.Sum256(body(tx)) // 教学标识，不是 Bitcoin txid。
	for i := range tx.Outputs {
		if _, exists := v[Outpoint{id, uint32(i)}]; exists {
			return 0, fmt.Errorf("output identifier already exists")
		}
	}
	// 全部验证结束后才提交；失败路径不会部分消费输入。
	for _, in := range tx.Inputs {
		delete(v, in)
	}
	for i, out := range tx.Outputs {
		v[Outpoint{id, uint32(i)}] = out
	}
	return inSum - outSum, nil
}

func key() ([32]byte, ed25519.PrivateKey) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	var p [32]byte
	copy(p[:], pub)
	return p, priv
}
func sign(tx *Tx, v View, priv ed25519.PrivateKey) {
	tx.Sigs = make([][]byte, len(tx.Inputs))
	for i := range tx.Inputs {
		d, err := digest(*tx, v, i)
		if err != nil {
			panic(err)
		}
		tx.Sigs[i] = ed25519.Sign(priv, d[:])
	}
}
func mustReject(label string, v View, tx Tx) {
	before := len(v)
	_, err := apply(v, tx)
	if err == nil || len(v) != before {
		panic("expected rejection without state change")
	}
	fmt.Println(label+":", err)
}

func main() {
	alice, aliceKey := key()
	bob, _ := key()
	start := Outpoint{ID: sha256.Sum256([]byte("teaching initial coin"))}
	v := View{start: Coin{Value: 1000, Owner: alice}}
	tx := Tx{Inputs: []Outpoint{start}, Outputs: []Coin{{600, bob}, {300, alice}}}
	sign(&tx, v, aliceKey)
	changed := tx
	changed.Outputs = append([]Coin(nil), tx.Outputs...)
	changed.Outputs[0].Value = 601
	mustReject("tampered output", v, changed)
	duplicate := Tx{Inputs: []Outpoint{start, start}, Outputs: []Coin{{1800, bob}}}
	sign(&duplicate, v, aliceKey)
	mustReject("duplicate input", v, duplicate)
	overspend := Tx{Inputs: []Outpoint{start}, Outputs: []Coin{{1001, bob}}}
	sign(&overspend, v, aliceKey)
	mustReject("signed overspend", v, overspend)
	outOfRange := Tx{Inputs: []Outpoint{start}, Outputs: []Coin{{^uint64(0), bob}}}
	sign(&outOfRange, v, aliceKey)
	mustReject("money range", v, outOfRange)
	fee, err := apply(v, tx)
	if err != nil || fee != 100 || len(v) != 2 {
		panic("valid transfer failed")
	}
	fmt.Printf("accepted: fee=%d, new UTXOs=%d\n", fee, len(v))
	mustReject("replay", v, tx)
}
