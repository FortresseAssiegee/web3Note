// 运行示例：PERF-23
package main

import "fmt"

func mark(graph [][]int, roots []int) []bool {
	live := make([]bool, len(graph))
	queue := append([]int(nil), roots...)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if live[current] {
			continue
		}
		live[current] = true
		queue = append(queue, graph[current]...)
	}
	return live
}

func main() {
	graph := [][]int{{1}, {2}, {1}, nil}
	live := mark(graph, []int{0})
	count := 0
	for _, reachable := range live {
		if reachable {
			count++
		}
	}
	fmt.Println("reachable:", count)
	fmt.Println("unreachable:", len(graph)-count)
}
