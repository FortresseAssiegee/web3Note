// 运行示例：PRJ-05
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("symbol") != "BTCUSDT" {
			http.Error(w, "bad symbol", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"symbol":"BTCUSDT","price":"60000.10"}`)
	}))
	defer server.Close()
	res, err := server.Client().Get(server.URL + "?symbol=BTCUSDT")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		panic(res.Status)
	}
	var q struct {
		Price string `json:"price"`
	}
	if err := json.NewDecoder(res.Body).Decode(&q); err != nil {
		panic(err)
	}
	fmt.Println(res.StatusCode, q.Price)
}
