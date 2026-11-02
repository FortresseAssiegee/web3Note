// 运行示例：PERF-13
package main

import "fmt"

func newCounter(start int) *int {
	counter := start
	return &counter
}

func main() {
	value := newCounter(41)
	*value++
	fmt.Println(*value)
}
