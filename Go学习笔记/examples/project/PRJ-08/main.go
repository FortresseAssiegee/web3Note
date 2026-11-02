// 运行示例：PRJ-08
package main

import (
	"fmt"
	"net/http"
)

func retryable(method string, status int) bool {
	if method != http.MethodGet {
		return false
	}
	return status == http.StatusTooManyRequests ||
		status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

func main() {
	fmt.Println(retryable("GET", 503))
	fmt.Println(retryable("GET", 400))
	fmt.Println(retryable("POST", 503))
}
