// 运行示例：ENV-12
package main

import (
	"fmt"
	"go/format"
)

func main() {
	source := []byte("package config\nconst Port=8080\n")
	formatted, err := format.Source(source)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(formatted))
}
