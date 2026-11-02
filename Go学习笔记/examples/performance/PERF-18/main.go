// 运行示例：PERF-18
package main

import "fmt"

func trySubmit(queue chan<- int, value int) bool {
	select {
	case queue <- value:
		return true
	default:
		return false
	}
}

func main() {
	queue := make(chan int, 2)
	fmt.Println(trySubmit(queue, 1))
	fmt.Println(trySubmit(queue, 2))
	fmt.Println(trySubmit(queue, 3))
	fmt.Println("queued:", len(queue))
}
