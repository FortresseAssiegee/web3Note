// 运行示例：PRJ-24
package main

import (
	"context"
	"fmt"
)

func worker(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	<-ctx.Done()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go worker(ctx, done)
	cancel()
	<-done
	fmt.Println("worker stopped")
}
