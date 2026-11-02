// 受限教学栈机：不是 Bitcoin Core Script 解释器，不能用于链上验证。
package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

type Instruction struct {
	Op   string
	Data []byte
}

func push(b []byte) Instruction { return Instruction{"PUSH", b} }
func op(s string) Instruction   { return Instruction{Op: s} }
func number(n int64) []byte {
	if n == 0 {
		return nil
	}
	negative := n < 0
	if negative {
		n = -n // 本例只编码受限整数，调用者不传 MinInt64。
	}
	var b []byte
	for n > 0 {
		b = append(b, byte(n))
		n >>= 8
	}
	if b[len(b)-1]&0x80 != 0 {
		b = append(b, 0)
	}
	if negative {
		b[len(b)-1] |= 0x80
	}
	return b
}
func decode(b []byte) (int64, error) {
	// 为了简明，本教学模型的数值操作只接受最小编码的 4 字节数。
	if len(b) > 4 {
		return 0, fmt.Errorf("number exceeds teaching limit")
	}
	if len(b) == 0 {
		return 0, nil
	}
	if b[len(b)-1]&0x7f == 0 && (len(b) == 1 || b[len(b)-2]&0x80 == 0) {
		return 0, fmt.Errorf("non-minimal number")
	}
	var n int64
	for i, v := range b {
		if i == len(b)-1 {
			v &= 0x7f
		}
		n |= int64(v) << (8 * i)
	}
	if b[len(b)-1]&0x80 != 0 {
		n = -n
	}
	return n, nil
}
func truth(b []byte) bool {
	for i, v := range b {
		if v != 0 {
			return !(i == len(b)-1 && v == 0x80) // 负零也是 false。
		}
	}
	return false
}
func boolean(v bool) []byte {
	if v {
		return []byte{1}
	}
	return nil
}
func run(program []Instruction) (bool, error) {
	var stack [][]byte
	pop := func() ([]byte, error) {
		if len(stack) == 0 {
			return nil, fmt.Errorf("stack underflow")
		}
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return v, nil
	}
	for _, ins := range program {
		switch ins.Op {
		case "PUSH":
			if len(ins.Data) > 520 {
				return false, fmt.Errorf("element too large")
			}
			stack = append(stack, append([]byte(nil), ins.Data...))
		case "2DUP":
			if len(stack) < 2 {
				return false, fmt.Errorf("stack underflow")
			}
			stack = append(stack, append([]byte(nil), stack[len(stack)-2]...), append([]byte(nil), stack[len(stack)-1]...))
		case "SWAP":
			if len(stack) < 2 {
				return false, fmt.Errorf("stack underflow")
			}
			n := len(stack)
			stack[n-1], stack[n-2] = stack[n-2], stack[n-1]
		case "ADD", "EQUAL":
			a, err := pop()
			if err != nil {
				return false, err
			}
			b, err := pop()
			if err != nil {
				return false, err
			}
			if ins.Op == "EQUAL" {
				stack = append(stack, boolean(bytes.Equal(a, b)))
			} else {
				x, err := decode(a)
				if err != nil {
					return false, err
				}
				y, err := decode(b)
				if err != nil {
					return false, err
				}
				stack = append(stack, number(x+y))
			}
		case "NOT", "VERIFY", "SHA256":
			a, err := pop()
			if err != nil {
				return false, err
			}
			switch ins.Op {
			case "NOT":
				n, err := decode(a)
				if err != nil {
					return false, err
				}
				stack = append(stack, boolean(n == 0))
			case "VERIFY":
				if !truth(a) {
					return false, fmt.Errorf("VERIFY failed")
				}
			case "SHA256":
				h := sha256.Sum256(a)
				stack = append(stack, h[:])
			}
		default:
			return false, fmt.Errorf("unsupported opcode %q", ins.Op)
		}
		if len(stack) > 1000 {
			return false, fmt.Errorf("stack too deep")
		}
	}
	// 本模型主动要求 clean stack；不是所有历史 Script 环境的通用规则。
	return len(stack) == 1 && truth(stack[0]), nil
}
func check(label string, p []Instruction, want bool) {
	ok, err := run(p)
	if (ok && err == nil) != want {
		panic(label)
	}
	fmt.Printf("%s: accepted=%v, error=%v\n", label, ok && err == nil, err)
}
func main() {
	check("3+4=7", []Instruction{push(number(3)), push(number(4)), op("ADD"), push(number(7)), op("EQUAL")}, true)
	secret := []byte("classroom preimage")
	h := sha256.Sum256(secret)
	check("hash preimage", []Instruction{push(secret), op("SHA256"), push(h[:]), op("EQUAL")}, true)
	check("wrong preimage", []Instruction{push([]byte("wrong")), op("SHA256"), push(h[:]), op("EQUAL")}, false)
	check("x=y is not a collision", []Instruction{push(secret), push(secret), op("2DUP"), op("EQUAL"), op("NOT"), op("VERIFY"), op("SHA256"), op("SWAP"), op("SHA256"), op("EQUAL")}, false)
	// 独立演示截断到 8 位后的碰撞；VM 上面一直使用完整 SHA-256。
	seen := map[byte]string{}
	for i := 0; ; i++ {
		s := fmt.Sprintf("toy-%d", i)
		d := sha256.Sum256([]byte(s))
		if previous, ok := seen[d[0]]; ok {
			p := sha256.Sum256([]byte(previous))
			fmt.Printf("8-bit toy collision: %q / %q; prefix=%02x; full hashes equal=%v\n", previous, s, d[0], p == d)
			break
		}
		seen[d[0]] = s
	}
}
