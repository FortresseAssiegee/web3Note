// BIP173/BIP350 地址教学实现；不生成私钥、不查询余额、不发送交易。
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

const alphabet = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
const bech32m uint32 = 0x2bc830a3

func polymod(v []byte) uint32 {
	gen := [5]uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := uint32(1)
	for _, x := range v {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ uint32(x)
		for i, g := range gen {
			if (top>>i)&1 != 0 {
				chk ^= g
			}
		}
	}
	return chk
}
func expand(hrp string) []byte {
	v := make([]byte, 0, 2*len(hrp)+1)
	for i := range hrp {
		v = append(v, hrp[i]>>5)
	}
	v = append(v, 0)
	for i := range hrp {
		v = append(v, hrp[i]&31)
	}
	return v
}
func convert(data []byte, from, to uint, pad bool) ([]byte, error) {
	var acc uint32
	var bits uint
	var out []byte
	mask := uint32((1 << to) - 1)
	maxAcc := uint32((1 << (from + to - 1)) - 1)
	for _, v := range data {
		if uint32(v)>>from != 0 {
			return nil, fmt.Errorf("input exceeds bit width")
		}
		acc = ((acc << from) | uint32(v)) & maxAcc
		bits += from
		for bits >= to {
			bits -= to
			out = append(out, byte((acc>>bits)&mask))
		}
	}
	if pad {
		if bits > 0 {
			out = append(out, byte((acc<<(to-bits))&mask))
		}
	} else if bits >= from || ((acc<<(to-bits))&mask) != 0 {
		return nil, fmt.Errorf("non-canonical padding")
	}
	return out, nil
}
func validProgram(version byte, program []byte) bool {
	return version <= 16 && len(program) >= 2 && len(program) <= 40 &&
		(version != 0 || len(program) == 20 || len(program) == 32)
}
func network(hrp string) bool { return hrp == "bc" || hrp == "tb" || hrp == "bcrt" }
func encodeRaw(hrp string, data []byte, constant uint32) string {
	values := append(expand(hrp), data...)
	values = append(values, make([]byte, 6)...)
	mod := polymod(values) ^ constant
	var b strings.Builder
	b.WriteString(hrp)
	b.WriteByte('1')
	for _, v := range data {
		b.WriteByte(alphabet[v])
	}
	for i := 0; i < 6; i++ {
		b.WriteByte(alphabet[(mod>>uint(5*(5-i)))&31])
	}
	return b.String()
}
func encode(hrp string, version byte, program []byte) (string, error) {
	if !network(hrp) || !validProgram(version, program) {
		return "", fmt.Errorf("invalid network or witness program")
	}
	data, err := convert(program, 8, 5, true)
	if err != nil {
		return "", err
	}
	constant := uint32(1)
	if version != 0 {
		constant = bech32m
	}
	return encodeRaw(hrp, append([]byte{version}, data...), constant), nil
}
func decode(addr, expectedHRP string) (byte, []byte, error) {
	if !network(expectedHRP) || len(addr) < 8 || len(addr) > 90 {
		return 0, nil, fmt.Errorf("invalid length or expected network")
	}
	if addr != strings.ToLower(addr) && addr != strings.ToUpper(addr) {
		return 0, nil, fmt.Errorf("mixed case")
	}
	for i := range addr {
		if addr[i] < 33 || addr[i] > 126 {
			return 0, nil, fmt.Errorf("invalid character range")
		}
	}
	addr = strings.ToLower(addr)
	separator := strings.LastIndexByte(addr, '1')
	if separator < 1 || separator+7 >= len(addr) || addr[:separator] != expectedHRP {
		return 0, nil, fmt.Errorf("wrong HRP or missing data")
	}
	data := make([]byte, len(addr)-separator-1)
	for i, c := range []byte(addr[separator+1:]) {
		n := strings.IndexByte(alphabet, c)
		if n < 0 {
			return 0, nil, fmt.Errorf("invalid data character")
		}
		data[i] = byte(n)
	}
	version := data[0]
	constant := uint32(1)
	if version != 0 {
		constant = bech32m
	}
	if polymod(append(expand(expectedHRP), data...)) != constant {
		return 0, nil, fmt.Errorf("checksum or encoding/version mismatch")
	}
	program, err := convert(data[1:len(data)-6], 5, 8, false)
	if err != nil || !validProgram(version, program) {
		return 0, nil, fmt.Errorf("invalid witness program or padding")
	}
	return version, program, nil
}
func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
func main() {
	// 来自 BIP350 的公开测试向量，不是钱包收款地址建议。
	vectors := []struct{ addr, hrp, script string }{
		{"BC1QW508D6QEJXTDG4Y5R3ZARVARY0C5XW7KV8F3T4", "bc", "0014751e76e8199196d454941c45d1b3a323f1433bd6"},
		{"tb1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3q0sl5k7", "tb", "00201863143c14c5166804bd19203356da136c985678cd4d27a1b8c6329604903262"},
		{"bc1p0xlxvlhemja6c4dqv22uapctqupfhlxm9h8z3k2e72q4k9hcz7vqzk5jj0", "bc", "512079be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"},
		{"BC1SW50QGDZ25J", "bc", "6002751e"},
	}
	for _, test := range vectors {
		version, program, err := decode(test.addr, test.hrp)
		if err != nil {
			panic(err)
		}
		opcode := byte(0)
		if version != 0 {
			opcode = 0x50 + version
		}
		script := append([]byte{opcode, byte(len(program))}, program...)
		roundtrip, err := encode(test.hrp, version, program)
		if err != nil || !bytes.Equal(script, mustHex(test.script)) || roundtrip != strings.ToLower(test.addr) {
			panic("official vector mismatch")
		}
		fmt.Printf("official vector: v=%d length=%d script=%x\n", version, len(program), script)
	}
	program := mustHex("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798")
	five, _ := convert(program, 8, 5, true)
	wrongEncoding := encodeRaw("bc", append([]byte{1}, five...), 1)
	badPadding := append([]byte(nil), five...)
	badPadding[len(badPadding)-1] |= 1 // 32 字节程序的最后 4 个填充位必须为零。
	wrongPadding := encodeRaw("bc", append([]byte{1}, badPadding...), bech32m)
	badCases := []struct{ address, network string }{
		{wrongEncoding, "bc"},
		{vectors[2].addr, "tb"},
		{"bC" + vectors[2].addr[2:], "bc"},
		{vectors[2].addr[:len(vectors[2].addr)-1] + "q", "bc"},
		{wrongPadding, "bc"},
	}
	for _, test := range badCases {
		if _, _, err := decode(test.address, test.network); err == nil {
			panic("invalid address accepted")
		} else {
			fmt.Println("rejected:", err)
		}
	}
}
