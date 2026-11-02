// 运行示例：PERF-17
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	var count atomic.Int64
	var wg sync.WaitGroup
	for worker := 0; worker < 2; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				count.Add(1)
			}
		}()
	}
	wg.Wait()
	fmt.Println("count:", count.Load())
}
