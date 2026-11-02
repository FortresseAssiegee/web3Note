// 运行示例：ENV-15
package main

import (
	"fmt"
	"go/build/constraint"
)

func main() {
	expression, err := constraint.Parse("//go:build linux && amd64 && !integration")
	if err != nil {
		panic(err)
	}
	linuxTags := map[string]bool{"linux": true, "amd64": true}
	windowsTags := map[string]bool{"windows": true, "arm64": true}
	fmt.Printf("linux-amd64=%v\n", expression.Eval(func(tag string) bool {
		return linuxTags[tag]
	}))
	fmt.Printf("windows-arm64=%v\n", expression.Eval(func(tag string) bool {
		return windowsTags[tag]
	}))
}
