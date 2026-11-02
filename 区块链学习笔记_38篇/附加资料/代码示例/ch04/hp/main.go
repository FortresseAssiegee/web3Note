package main

import (
	"bytes"
	"fmt"
)

// Nibbles contains only digits 0..15. Leaf status is passed separately.
func encodeHP(path []byte, isLeaf bool) ([]byte, error) {
	for _, n := range path {
		if n > 15 {
			return nil, fmt.Errorf("invalid nibble: %d", n)
		}
	}
	flag := byte(0)
	if isLeaf {
		flag = 2
	}
	odd := len(path)%2 == 1
	out := make([]byte, 1, 1+len(path)/2)
	if odd {
		out[0] = (flag+1)<<4 | path[0]
		path = path[1:]
	} else {
		out[0] = flag << 4
	}
	for i := 0; i < len(path); i += 2 {
		out = append(out, path[i]<<4|path[i+1])
	}
	return out, nil
}

func decodeHP(encoded []byte) ([]byte, bool, error) {
	if len(encoded) == 0 {
		return nil, false, fmt.Errorf("empty compact encoding")
	}
	flag := encoded[0] >> 4
	if flag > 3 {
		return nil, false, fmt.Errorf("invalid HP flag")
	}
	odd := flag&1 == 1
	if !odd && encoded[0]&15 != 0 {
		return nil, false, fmt.Errorf("nonzero even-path padding")
	}
	var path []byte
	if odd {
		path = append(path, encoded[0]&15)
	}
	for _, b := range encoded[1:] {
		path = append(path, b>>4, b&15)
	}
	return path, flag&2 != 0, nil
}

func bytesToNibbles(raw []byte) []byte {
	path := make([]byte, 0, 2*len(raw))
	for _, b := range raw {
		path = append(path, b>>4, b&15)
	}
	return path
}

func main() {
	cases := []struct {
		name string
		path []byte
		leaf bool
	}{
		{"extension-even", []byte{0xa, 0xb}, false},
		{"extension-odd", []byte{0xa, 0xb, 0xc}, false},
		{"leaf-even", []byte{0xa, 0xb}, true},
		{"leaf-odd", []byte{0xa, 0xb, 0xc}, true},
		{"leaf-empty", nil, true},
	}
	for _, c := range cases {
		encoded, err := encodeHP(c.path, c.leaf)
		if err != nil {
			panic(err)
		}
		decoded, leaf, err := decodeHP(encoded)
		if err != nil || !bytes.Equal(decoded, c.path) || leaf != c.leaf {
			panic("round trip failed")
		}
		fmt.Printf("%s: %x roundtrip=true\n", c.name, encoded)
	}
	fmt.Printf("raw ab -> nibbles %v\n", bytesToNibbles([]byte{0xab}))
	fmt.Printf("ASCII ab -> nibbles %v\n", bytesToNibbles([]byte("ab")))
	_, _, err := decodeHP([]byte{0x21})
	fmt.Println("reject bad padding:", err != nil)
	_, err = encodeHP([]byte{16}, true)
	fmt.Println("reject nibble 16:", err != nil)
}
