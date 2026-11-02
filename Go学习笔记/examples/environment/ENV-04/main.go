// 运行示例：ENV-04
package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	name := filepath.Join("config", "..", "assets", "app.json")
	fmt.Println(filepath.ToSlash(filepath.Clean(name)))
	fmt.Println(filepath.Base(name))
	fmt.Println(filepath.Ext(name))
}
