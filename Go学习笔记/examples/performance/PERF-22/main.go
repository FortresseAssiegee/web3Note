// 运行示例：PERF-22
package main

import (
	"fmt"
	"runtime/pprof"
)

func main() {
	leases := pprof.NewProfile("lesson.open.leases")
	token := new(int)
	leases.Add(token, 0)
	fmt.Println("active:", leases.Count())
	leases.Remove(token)
	fmt.Println("active:", leases.Count())
}
