// 运行示例：PERF-20
package main

import (
	"bytes"
	"context"
	"fmt"
	"runtime/trace"
)

func main() {
	var recording bytes.Buffer
	if err := trace.Start(&recording); err != nil {
		panic(err)
	}
	ctx, task := trace.NewTask(context.Background(), "quote-request")
	result := make(chan int, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		trace.WithRegion(ctx, "calculate", func() {
			result <- 1 + 2 + 3
		})
	}()
	var total int
	trace.WithRegion(ctx, "wait-result", func() { total = <-result })
	<-done
	task.End()
	trace.Stop()
	fmt.Println("result:", total)
	fmt.Println("trace written:", recording.Len() > 0)
}
