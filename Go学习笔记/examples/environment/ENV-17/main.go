// 运行示例：ENV-17
package main

import (
	"fmt"
	"time"
)

func main() {
	for _, raw := range []string{"500ms", "2s", "5"} {
		duration, err := time.ParseDuration(raw)
		if err != nil {
			fmt.Printf("%q rejected=true\n", raw)
			continue
		}
		fmt.Printf("%q milliseconds=%d\n", raw, duration.Milliseconds())
	}
}
