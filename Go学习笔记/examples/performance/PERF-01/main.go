// 运行示例：PERF-01
package main

import (
	"fmt"
	"math"
	"slices"
)

func nearestRank(values []int, percentile float64) int {
	if len(values) == 0 || percentile <= 0 || percentile > 1 {
		panic("invalid input")
	}
	ordered := slices.Clone(values)
	slices.Sort(ordered)
	return ordered[int(math.Ceil(percentile*float64(len(ordered))))-1]
}

func main() {
	samples := []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 1000}
	total := 0
	for _, x := range samples {
		total += x
	}
	fmt.Println("mean:", total/len(samples))
	fmt.Println("p50:", nearestRank(samples, 0.50))
	fmt.Println("p95:", nearestRank(samples, 0.95))
}
