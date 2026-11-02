// 运行示例：PRJ-14
package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Row struct {
	Exchange       string `json:"exchange"`
	Symbol         string `json:"symbol"`
	ObservationID  string `json:"observation_id"`
	ObservedMillis int64  `json:"observed_unix_ms"`
	Price          string `json:"price"`
}

func main() {
	r := Row{"demo-binance-spot", "BTCUSDT", "obs-a", time.Unix(1, 0).UnixMilli(), "60000.10"}
	data, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
