// 运行示例：PERF-03
package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http/httptest"
)

func main() {
	requests := expvar.NewInt("lesson_requests_total")
	requests.Add(3)
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/debug/vars", nil)
	expvar.Handler().ServeHTTP(res, req)
	var values map[string]json.RawMessage
	if err := json.Unmarshal(res.Body.Bytes(), &values); err != nil {
		panic(err)
	}
	fmt.Println("requests:", string(values["lesson_requests_total"]))
}
