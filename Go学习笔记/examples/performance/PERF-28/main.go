// 运行示例：PERF-28
package main

import "fmt"

func main() {
	values := []*int{new(int), new(int), new(int)}
	kept := values[:1]
	clear(values[1:])
	fmt.Println("visible length:", len(kept))
	fmt.Println("tail cleared:", values[1] == nil && values[2] == nil)
}
