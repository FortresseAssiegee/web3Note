// 运行示例：PERF-25
package main

import (
	"fmt"
	"runtime/debug"
)

func main() {
	oldPercent := debug.SetGCPercent(150)
	defer debug.SetGCPercent(oldPercent)
	oldLimit := debug.SetMemoryLimit(64 << 20)
	defer debug.SetMemoryLimit(oldLimit)

	currentPercent := debug.SetGCPercent(150)
	currentLimit := debug.SetMemoryLimit(-1)
	fmt.Println("GOGC:", currentPercent)
	fmt.Println("memory limit MiB:", currentLimit>>20)
}
