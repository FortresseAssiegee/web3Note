package main

import (
	"example.com/learning/left"
	"example.com/learning/right"
	"fmt"
)

func main() {
	fmt.Printf("left=%d right=%d\n", left.Value(), right.Value())
}
