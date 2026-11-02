// 运行示例：PRJ-20
package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	err := fmt.Errorf("fetch BTCUSDT: %w", context.DeadlineExceeded)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))
	fmt.Println(errors.Is(err, context.Canceled))
}
