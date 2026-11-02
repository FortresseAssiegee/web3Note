// 运行示例：PRJ-04
package main

import (
	"encoding/json"
	"fmt"
)

func valid(data string) bool {
	var v struct {
		Symbol string `json:"symbol"`
	}
	return json.Unmarshal([]byte(data), &v) == nil
}

func main() {
	fmt.Println(valid(`{"symbol":"BTCUSDT"}`))
	fmt.Println(valid(`{"symbol":"BTCUSDT"}{}`))
}
