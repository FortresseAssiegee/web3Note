// 运行示例：PRJ-10
package main

import (
	"fmt"
	"strings"
)

func key(exchange, symbol string) (string, error) {
	for _, part := range []string{exchange, symbol} {
		if part == "" || strings.ContainsAny(part, ":\r\n ") {
			return "", fmt.Errorf("invalid key component")
		}
	}
	return "quotes:v1:" + exchange + ":" + symbol, nil
}

func main() {
	k, err := key("binance-spot", "BTCUSDT")
	if err != nil {
		panic(err)
	}
	fmt.Println(k)
	_, err = key("binance:spot", "BTCUSDT")
	fmt.Println(err != nil)
}
