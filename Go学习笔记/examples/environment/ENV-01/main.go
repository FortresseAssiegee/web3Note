// 运行示例：ENV-01
package main

import "fmt"

func main() {
	checks := []string{"source", "compiler", "runtime"}
	for i, check := range checks {
		fmt.Printf("%d: %s ready\n", i+1, check)
	}
}
