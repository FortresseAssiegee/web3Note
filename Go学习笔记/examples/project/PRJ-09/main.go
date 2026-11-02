// 运行示例：PRJ-09
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func retryAfter(raw string, now time.Time) (time.Duration, bool) {
	if n, err := strconv.Atoi(raw); err == nil {
		if n < 0 || n > 60 {
			return 0, false
		}
		return time.Duration(n) * time.Second, true
	}
	at, err := http.ParseTime(raw)
	if err != nil {
		return 0, false
	}
	d := at.Sub(now)
	if d < 0 {
		d = 0
	}
	if d > time.Minute {
		return 0, false
	}
	return d, true
}

func main() {
	now := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	d, ok := retryAfter("10", now)
	fmt.Println(d, ok)
	d, ok = retryAfter(now.Add(20*time.Second).Format(http.TimeFormat), now)
	fmt.Println(d, ok)
	_, ok = retryAfter("-1", now)
	fmt.Println(ok)
}
