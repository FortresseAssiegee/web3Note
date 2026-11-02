// 运行示例：PERF-24
package main

import "fmt"

func approximateGoalMiB(live, roots, gogc float64) float64 {
	return live + (live+roots)*gogc/100
}

func main() {
	for _, value := range []float64{100, 200, 50} {
		fmt.Printf("GOGC=%.0f goal=%.1f MiB\n", value, approximateGoalMiB(4, 1, value))
	}
}
