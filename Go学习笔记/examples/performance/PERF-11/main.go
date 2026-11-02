// 运行示例：PERF-11
package main

import (
	"fmt"
	"strconv"
	"strings"
)

func formatGeneric(symbol string, count int) string {
	return fmt.Sprintf("%s:%d", symbol, count)
}

func formatSpecific(symbol string, count int) string {
	var b strings.Builder
	b.Grow(len(symbol) + 24)
	b.WriteString(symbol)
	b.WriteByte(':')
	b.WriteString(strconv.Itoa(count))
	return b.String()
}

func main() {
	a := formatGeneric("BTCUSDT", 12)
	b := formatSpecific("BTCUSDT", 12)
	fmt.Println(a)
	fmt.Println("equal:", a == b)
}
