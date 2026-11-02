// 运行示例：ENV-13
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	source := `package sample
import "fmt"
func run() { fmt.Println("ok") }
func main() { run() }
`
	file, err := parser.ParseFile(token.NewFileSet(), "sample.go", source, 0)
	if err != nil {
		panic(err)
	}
	var names []string
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			names = append(names, function.Name.Name)
		}
	}
	fmt.Printf("package=%s declarations=%d\n", file.Name.Name, len(file.Decls))
	fmt.Printf("functions=%v\n", names)
}
