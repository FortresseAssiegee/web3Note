// 运行示例：PERF-10
package main

import (
	"fmt"
	"runtime"
	"runtime/metrics"
)

func cycles() uint64 {
	samples := []metrics.Sample{{Name: "/gc/cycles/total:gc-cycles"}}
	metrics.Read(samples)
	if samples[0].Value.Kind() != metrics.KindUint64 {
		panic("metric unavailable")
	}
	return samples[0].Value.Uint64()
}

func main() {
	before := cycles()
	runtime.GC()
	after := cycles()
	fmt.Println("cycle observed:", after > before)
}
