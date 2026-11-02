// 运行示例：PRJ-13
package main

import "fmt"

type Value struct {
	Seq   int64
	Price string
}

func apply(current *Value, next Value) bool {
	if next.Seq <= current.Seq {
		return false
	}
	*current = next
	return true
}

func main() {
	current := Value{Seq: 10, Price: "100.00"}
	fmt.Println(apply(&current, Value{Seq: 12, Price: "102.00"}))
	fmt.Println(apply(&current, Value{Seq: 11, Price: "101.00"}))
	fmt.Println(current.Seq, current.Price)
}
