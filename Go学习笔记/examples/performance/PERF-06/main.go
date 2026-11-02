// 运行示例：PERF-06
package main

import (
	"context"
	"fmt"
)

func integers(ctx context.Context) (<-chan int, <-chan struct{}) {
	out := make(chan int)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer close(out)
		for n := 0; ; n++ {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, done
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	values, done := integers(ctx)
	for i := 0; i < 3; i++ {
		fmt.Println(<-values)
	}
	cancel()
	<-done
	fmt.Println("producer stopped")
}
