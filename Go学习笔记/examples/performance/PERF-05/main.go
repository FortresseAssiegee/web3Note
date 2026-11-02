// 运行示例：PERF-05
package main

import (
	"context"
	"fmt"
	"runtime/pprof"
)

func main() {
	ctx := context.Background()
	pprof.Do(ctx, pprof.Labels("route", "quote"), func(labeled context.Context) {
		route, ok := pprof.Label(labeled, "route")
		fmt.Println(route, ok)
	})
	_, exists := pprof.Label(ctx, "route")
	fmt.Println("outside:", exists)
}
