// 运行示例：PRJ-16
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func label(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "quotes_" + hex.EncodeToString(sum[:])
}

func main() {
	a := []byte("row-a\nrow-b\n")
	b := []byte("row-a\nrow-c\n")
	fmt.Println(label(a) == label(a))
	fmt.Println(label(a) != label(b))
	fmt.Println(len(label(a)))
}
