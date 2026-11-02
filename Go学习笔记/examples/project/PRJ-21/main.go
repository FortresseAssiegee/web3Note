// 运行示例：PRJ-21
package main

import (
	"bytes"
	"fmt"
)

func main() {
	stored := []byte("60000.10")
	returned := bytes.Clone(stored)
	returned[0] = '9'
	fmt.Println(string(stored))
	fmt.Println(string(returned))
}
