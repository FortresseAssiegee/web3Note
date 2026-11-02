// 运行示例：PERF-02
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/pprof"
)

func main() {
	business := http.NewServeMux()
	business.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})
	diagnostics := http.NewServeMux()
	diagnostics.HandleFunc("GET /debug/pprof/", pprof.Index)
	diagnostics.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	diagnostics.HandleFunc("GET /debug/pprof/trace", pprof.Trace)

	for _, handler := range []http.Handler{business, diagnostics} {
		req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		fmt.Println(res.Code)
	}
}
