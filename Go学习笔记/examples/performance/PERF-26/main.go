// 运行示例：PERF-26
package main

import "fmt"

func shade(colors map[string]string, object string) {
	if object != "" && colors[object] == "white" {
		colors[object] = "gray"
	}
}

func writePointerModel(colors map[string]string, slot *string, next string, grayStack bool) {
	shade(colors, *slot)
	if grayStack {
		shade(colors, next)
	}
	*slot = next
}

func main() {
	colors := map[string]string{"old": "white", "new": "white"}
	slot := "old"
	writePointerModel(colors, &slot, "new", true)
	fmt.Println("old:", colors["old"])
	fmt.Println("new:", colors["new"])
	fmt.Println("slot:", slot)
}
