// 运行示例：PRJ-18
package main

import "fmt"

func main() {
	query := "SELECT price FROM observations WHERE symbol = ? LIMIT ?"
	args := []any{"BTCUSDT", 20}
	fmt.Println(query)
	fmt.Printf("%s %d\n", args[0], args[1])
}
