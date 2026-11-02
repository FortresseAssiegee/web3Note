// 运行示例：ENV-07
package main

import (
	"fmt"
	"strconv"
)

func portValue(raw string) (int, error) {
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid port %q: %w", raw, err)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port %d is out of range", port)
	}
	return port, nil
}

func main() {
	for _, raw := range []string{"8080", "0", "65536", "eighty"} {
		port, err := portValue(raw)
		if err != nil {
			fmt.Printf("%q rejected\n", raw)
			continue
		}
		fmt.Printf("%q accepted=%d\n", raw, port)
	}
}
