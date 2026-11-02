// 只验证 Taproot 公钥承诺和控制块；不执行脚本、不签交易、不广播。
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

func tagged(tag string, parts ...[]byte) [32]byte {
	t := sha256.Sum256([]byte(tag))
	h := sha256.New()
	h.Write(t[:])
	h.Write(t[:])
	for _, part := range parts {
		h.Write(part)
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}
func compactSize(n uint64) []byte {
	var b bytes.Buffer
	switch {
	case n < 253:
		b.WriteByte(byte(n))
	case n <= 0xffff:
		b.WriteByte(253)
		_ = binary.Write(&b, binary.LittleEndian, uint16(n))
	case n <= 0xffffffff:
		b.WriteByte(254)
		_ = binary.Write(&b, binary.LittleEndian, uint32(n))
	default:
		b.WriteByte(255)
		_ = binary.Write(&b, binary.LittleEndian, n)
	}
	return b.Bytes()
}
func leaf(script []byte) [32]byte {
	return tagged("TapLeaf", []byte{0xc0}, compactSize(uint64(len(script))), script)
}
func branch(a, b [32]byte) [32]byte {
	if bytes.Compare(a[:], b[:]) > 0 {
		a, b = b, a
	}
	return tagged("TapBranch", a[:], b[:])
}
func tweak(internal, root []byte) ([]byte, byte, error) {
	if len(internal) != 32 || (len(root) != 0 && len(root) != 32) {
		return nil, 0, fmt.Errorf("invalid key/root length")
	}
	pub, err := schnorr.ParsePubKey(internal) // lift_x：选择偶数 y。
	if err != nil {
		return nil, 0, err
	}
	h := tagged("TapTweak", internal, root)
	var scalar btcec.ModNScalar
	if scalar.SetByteSlice(h[:]) {
		return nil, 0, fmt.Errorf("tweak >= curve order")
	}
	var p, t, q btcec.JacobianPoint
	pub.AsJacobian(&p)
	btcec.ScalarBaseMultNonConst(&scalar, &t)
	btcec.AddNonConst(&p, &t, &q)
	if q.Z.IsZero() {
		return nil, 0, fmt.Errorf("point at infinity")
	}
	q.ToAffine()
	out := btcec.NewPublicKey(&q.X, &q.Y)
	return schnorr.SerializePubKey(out), out.SerializeCompressed()[0] & 1, nil
}
func verifyCommitment(output, script, control []byte) bool {
	if len(output) != 32 || len(control) < 33 || len(control) > 33+32*128 || (len(control)-33)%32 != 0 {
		return false
	}
	// 本例只处理当前 Tapscript 叶版本；不是完整的未来版本共识解释器。
	if control[0]&0xfe != 0xc0 {
		return false
	}
	root := leaf(script)
	for i := 33; i < len(control); i += 32 {
		var sibling [32]byte
		copy(sibling[:], control[i:i+32])
		root = branch(root, sibling)
	}
	x, parity, err := tweak(control[1:33], root[:])
	return err == nil && bytes.Equal(x, output) && parity == control[0]&1
}
func hx(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
func main() {
	// BIP341 wallet-test-vectors.json 第一组，无脚本树。
	internal := hx("d6889cb081036e0faefa3a35157ad71086b123b2b144b649798b494c300a961d")
	x, _, err := tweak(internal, nil)
	if err != nil || !bytes.Equal(x, hx("53a1f6e454df1aa2776a2814a721372d6258050de330b3c6d10ee8f4e0dda343")) {
		panic("official no-tree vector failed")
	}
	fmt.Printf("official no-tree output key: %x\n", x)
	// BIP341 第二组：单叶脚本及控制块，检验叶哈希和奇偶位。
	script := hx("20d85a959b0290bf19bb89ed43c916be835475d013da4b362117393e25a48229b8ac")
	control := hx("c1187791b6f712a8ea41c8ecdd0ee77fab3e85263b37e1ec18a3651926b3a6cf27")
	output := hx("147c9c57132f6e7ecddba9800bb0c4449251c92a1e60371ee77557b6620f3ea3")
	if !verifyCommitment(output, script, control) {
		panic("official single-leaf vector failed")
	}
	fmt.Println("official single-leaf commitment: true")
	// 自创双叶例子：复用公开测试公钥，不得向这些教学脚本付款。
	other := hx("20b617298552a72ade070667e86ca63b8f5789a9fe8731ef91202a91c9f3459007ac")
	a, b := leaf(script), leaf(other)
	root := branch(a, b)
	x, parity, err := tweak(internal, root[:])
	if err != nil {
		panic(err)
	}
	proof := append([]byte{0xc0 | parity}, internal...)
	proof = append(proof, b[:]...)
	if !verifyCommitment(x, script, proof) {
		panic("two-leaf proof failed")
	}
	fmt.Printf("two-leaf proof: true; control bytes=%d\n", len(proof))
	badProof := append([]byte(nil), proof...)
	badProof[len(badProof)-1] ^= 1
	badParity := append([]byte(nil), proof...)
	badParity[0] ^= 1
	badScript := append([]byte(nil), script...)
	badScript[len(badScript)-1] ^= 1
	for label, ok := range map[string]bool{
		"changed sibling": verifyCommitment(x, script, badProof),
		"changed parity":  verifyCommitment(x, script, badParity),
		"changed script":  verifyCommitment(x, badScript, proof),
	} {
		if ok {
			panic("tampered commitment accepted")
		}
		fmt.Println(label+":", ok)
	}
}
