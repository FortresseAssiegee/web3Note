// 运行示例：ENV-05
package main

import (
	"fmt"
	"io/fs"
)

func main() {
	names := []string{"config/app.json", "../app.json", "/app.json", ".", ""}
	for _, name := range names {
		fmt.Printf("%q valid=%v\n", name, fs.ValidPath(name))
	}
}
