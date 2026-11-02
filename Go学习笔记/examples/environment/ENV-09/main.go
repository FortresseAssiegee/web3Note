// 运行示例：ENV-09
package main

import (
	"fmt"
	"os"
)

func main() {
	file, err := os.CreateTemp("", "go-env-note-*.txt")
	if err != nil {
		panic(err)
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err := file.WriteString("port=8080\n"); err != nil {
		file.Close()
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
	data, err := os.ReadFile(name)
	if err != nil {
		panic(err)
	}
	fmt.Printf("bytes=%d content=%q\n", len(data), data)
}
