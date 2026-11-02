// 运行示例：PRJ-07
package main

import (
	"context"
	"errors"
	"fmt"
)

func fetch(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fmt.Println(errors.Is(fetch(ctx), context.Canceled))
}
