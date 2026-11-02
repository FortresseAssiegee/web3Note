// 运行示例：PERF-12
package main

import (
	"fmt"
	"testing"
)

var retained []byte

func main() {
	allocations := testing.AllocsPerRun(100, func() {
		retained = make([]byte, 1024)
	})
	fmt.Println("allocation observed:", allocations >= 1)
	fmt.Println("retained bytes:", len(retained))
}
