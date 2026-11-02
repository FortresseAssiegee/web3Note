package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
)

// This is a byte trie with a custom commitment, NOT an Ethereum MPT.
type Node struct {
	children map[byte]*Node
	value    []byte
	present  bool
}

type Hash [32]byte

func newNode() *Node { return &Node{children: make(map[byte]*Node)} }

func (n *Node) put(key string, value []byte) {
	cur := n
	for i := 0; i < len(key); i++ {
		if cur.children[key[i]] == nil {
			cur.children[key[i]] = newNode()
		}
		cur = cur.children[key[i]]
	}
	cur.value = append([]byte(nil), value...)
	cur.present = true
}

func (n *Node) get(key string) ([]byte, bool) {
	cur := n
	for i := 0; i < len(key); i++ {
		cur = cur.children[key[i]]
		if cur == nil {
			return nil, false
		}
	}
	return append([]byte(nil), cur.value...), cur.present
}

func (n *Node) delete(key string) bool {
	if len(key) == 0 {
		wasPresent := n.present
		n.value, n.present = nil, false
		return wasPresent
	}
	child := n.children[key[0]]
	if child == nil {
		return false
	}
	removed := child.delete(key[1:])
	if !child.present && len(child.children) == 0 {
		delete(n.children, key[0])
	}
	return removed
}

func (n *Node) clone() *Node {
	x := newNode()
	x.value = append([]byte(nil), n.value...)
	x.present = n.present
	for k, c := range n.children {
		x.children[k] = c.clone()
	}
	return x
}

func (n *Node) hash() Hash {
	// Encoding: tag | presence | length:u64 | value | child-count:u16 | sorted children.
	b := []byte{0x54, 0}
	if n.present {
		b[1] = 1
	}
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(n.value)))
	b = append(b, length[:]...)
	b = append(b, n.value...)
	keys := make([]int, 0, len(n.children))
	for k := range n.children {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	var count [2]byte
	binary.BigEndian.PutUint16(count[:], uint16(len(keys)))
	b = append(b, count[:]...)
	for _, k := range keys {
		h := n.children[byte(k)].hash()
		b = append(b, byte(k))
		b = append(b, h[:]...)
	}
	return sha256.Sum256(b)
}

// A full in-memory snapshot models versions; this is not disk persistence.
func commit(n *Node, versions map[Hash]*Node) Hash {
	h := n.hash()
	versions[h] = n.clone()
	return h
}

func main() {
	t := newNode()
	t.put("car", []byte("10"))
	t.put("cart", []byte("20"))
	t.put("cat", []byte("30"))
	versions := make(map[Hash]*Node)
	r1 := commit(t, versions)
	v, ok := t.get("car")
	fmt.Printf("car=%s exists=%t\n", v, ok)
	t.put("car", []byte("11"))
	r2 := commit(t, versions)
	fmt.Println("update changes root:", r1 != r2)
	fmt.Println("delete car:", t.delete("car"))
	_, carExists := t.get("car")
	v, cartExists := t.get("cart")
	fmt.Printf("car exists=%t; cart=%s exists=%t\n", carExists, v, cartExists)
	v, ok = versions[r1].get("car")
	fmt.Printf("old snapshot car=%s exists=%t\n", v, ok)
	t.put("empty", nil)
	_, emptyExists := t.get("empty")
	fmt.Println("empty value differs from absent:", emptyExists)
}
