// 运行示例：ENV-11
package main

import (
	"crypto/sha256"
	"fmt"
)

func main() {
	trustedSample := []byte("go-lab\n")
	expected := sha256.Sum256(trustedSample)
	original := sha256.Sum256([]byte("go-lab\n"))
	changed := sha256.Sum256([]byte("go-lab!\n"))
	fmt.Printf("original matches=%v\n", original == expected)
	fmt.Printf("changed matches=%v\n", changed == expected)
}
