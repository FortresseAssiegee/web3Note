// 运行示例：PERF-09
package main

import (
	"fmt"
	"runtime/metrics"
)

func main() {
	samples := []metrics.Sample{
		{Name: "/gc/heap/allocs:bytes"},
		{Name: "/gc/heap/live:bytes"},
		{Name: "/sched/goroutines:goroutines"},
	}
	metrics.Read(samples)
	for _, sample := range samples {
		fmt.Println(sample.Name, sample.Value.Kind() == metrics.KindUint64)
	}
}
