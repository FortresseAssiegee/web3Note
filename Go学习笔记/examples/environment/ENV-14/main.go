// 运行示例：ENV-14
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	source := `package buildinfo
func Build() {}
func inspect() {}
type Config struct{}
func (Config) Validate() {}
`
	file, err := parser.ParseFile(token.NewFileSet(), "api.go", source, 0)
	if err != nil {
		panic(err)
	}
	var names []string
	ast.Inspect(file, func(node ast.Node) bool {
		function, ok := node.(*ast.FuncDecl)
		if ok && function.Recv == nil && ast.IsExported(function.Name.Name) {
			names = append(names, function.Name.Name)
		}
		return true
	})
	fmt.Printf("exported functions=%v\n", names)
}
