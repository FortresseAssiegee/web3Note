// 运行示例：ENV-18
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Module struct {
	Path    string
	Version string
	Main    bool
	Replace *Module
}

func main() {
	report := `{"Path":"example.com/app","Main":true}
{"Path":"example.com/number","Version":"v1.3.0"}
{"Path":"example.com/unit","Version":"v0.0.0","Replace":{"Path":"../unit"}}
`
	decoder := json.NewDecoder(strings.NewReader(report))
	for {
		var module Module
		err := decoder.Decode(&module)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		version := module.Version
		if module.Main {
			version = "(main)"
		}
		fmt.Printf("%s %s", module.Path, version)
		if module.Replace != nil {
			fmt.Printf(" => %s", module.Replace.Path)
		}
		fmt.Println()
	}
}
