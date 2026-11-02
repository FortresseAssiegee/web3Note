// 运行示例：PERF-08
package main

import (
	"fmt"
	"runtime/pprof"
)

func main() {
	fmt.Println("goroutine profile:", pprof.Lookup("goroutine") != nil)
	fmt.Println("goroutineleak profile:", pprof.Lookup("goroutineleak") != nil)
}
