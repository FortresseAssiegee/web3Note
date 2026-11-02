// 运行示例：PRJ-11
package main

import (
	"fmt"
	"time"
)

type Entry struct {
	Value   string
	Expires time.Time
}

func (e Entry) Get(now time.Time) (string, bool) {
	if !now.Before(e.Expires) {
		return "", false
	}
	return e.Value, true
}

func main() {
	start := time.Unix(100, 0)
	e := Entry{Value: "60000.10", Expires: start.Add(5 * time.Second)}
	v, ok := e.Get(start.Add(4 * time.Second))
	fmt.Println(v, ok)
	_, ok = e.Get(start.Add(5 * time.Second))
	fmt.Println(ok)
}
