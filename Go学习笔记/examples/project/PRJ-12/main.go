// 运行示例：PRJ-12
package main

import "fmt"

func enqueue(queue chan<- string, value string) bool {
	select {
	case queue <- value:
		return true
	default:
		return false
	}
}

func main() {
	queue := make(chan string, 1)
	fmt.Println(enqueue(queue, "BTCUSDT"))
	fmt.Println(enqueue(queue, "ETHUSDT"))
	fmt.Println(<-queue)
}
