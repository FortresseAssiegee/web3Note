package main

import (
	"example.com/learning/unit"
	"fmt"
)

func main() {
	text, err := unit.KiB(2048)
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
}
