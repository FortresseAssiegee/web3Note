// 运行示例：ENV-10
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir, err := os.MkdirTemp("", "go-env-dir-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	for _, name := range []string{"b.txt", "a.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0600); err != nil {
			panic(err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		fmt.Println(entry.Name())
	}
}
