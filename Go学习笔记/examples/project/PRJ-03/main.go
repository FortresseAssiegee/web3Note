// 运行示例：PRJ-03
package main

import (
	"fmt"
	"net/url"
)

func main() {
	u, err := url.Parse("https://example.test/api/v3/ticker/price")
	if err != nil {
		panic(err)
	}
	q := u.Query()
	q.Set("symbol", "BTCUSDT")
	q.Set("note", "a+b c")
	u.RawQuery = q.Encode()
	fmt.Println(u.Query().Get("symbol"))
	fmt.Println(u.RawQuery)
}
