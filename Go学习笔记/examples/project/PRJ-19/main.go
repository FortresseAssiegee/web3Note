// 运行示例：PRJ-19
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /quotes/{symbol}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r.PathValue("symbol"))
		fmt.Fprintln(w, r.URL.Query().Get("symbol"))
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/quotes/BTCUSDT?symbol=ETHUSDT", nil)
	mux.ServeHTTP(rec, req)
	fmt.Println(rec.Code)
	fmt.Print(rec.Body.String())
}
