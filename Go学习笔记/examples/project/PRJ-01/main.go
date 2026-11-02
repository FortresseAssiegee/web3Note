// 运行示例：PRJ-01
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	var q struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	err := json.Unmarshal([]byte(`{"symbol":"BTCUSDT","price":"60000.1000"}`), &q)
	if err != nil {
		panic(err)
	}
	fmt.Println(q.Symbol)
	fmt.Println(q.Price)
}
