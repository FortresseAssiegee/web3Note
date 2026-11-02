// 运行示例：ENV-08
package main

import (
	"fmt"
	"io/fs"
	"strings"
	"testing/fstest"
)

func loadText(fsys fs.FS, name string) (string, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func main() {
	files := fstest.MapFS{
		"config/app.txt": &fstest.MapFile{Data: []byte("port=8080\n"), Mode: 0644},
	}
	text, err := loadText(files, "config/app.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
}
