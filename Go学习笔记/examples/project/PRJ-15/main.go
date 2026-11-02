// 运行示例：PRJ-15
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func main() {
	rows := []struct {
		ID    string `json:"id"`
		Price string `json:"price"`
	}{
		{"a", "10.00"}, {"b", "11.00"},
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			panic(err)
		}
	}
	fmt.Print(buf.String())
	fmt.Println("rows:", bytes.Count(buf.Bytes(), []byte("\n")))
}
