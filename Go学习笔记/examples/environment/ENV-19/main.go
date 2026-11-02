// 运行示例：ENV-19
package main

import (
	"fmt"
	"testing/fstest"
)

func main() {
	files := fstest.MapFS{
		"config/app.txt": &fstest.MapFile{Data: []byte("ready\n"), Mode: 0644},
	}
	err := fstest.TestFS(files, "config/app.txt")
	fmt.Printf("fixture valid=%v\n", err == nil)
	err = fstest.TestFS(files, "config/missing.txt")
	fmt.Printf("missing detected=%v\n", err != nil)
}
