// 运行示例：PERF-14
package main

import (
	"fmt"
	"strings"
)

func prefixCopy(input string) string {
	return strings.Clone(input[:8])
}

func main() {
	large := strings.Repeat("ABCDEFGH", 1024)
	prefix := prefixCopy(large)
	fmt.Println(prefix)
	fmt.Println("length:", len(prefix))
}
