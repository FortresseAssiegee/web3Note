package main

import (
	"example.com/learning/hello/internal/greeting"
	"flag"
	"fmt"
	"os"
)

func main() {
	name := flag.String("name", "Gopher", "person to greet")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "unexpected positional arguments")
		os.Exit(2)
	}
	text, err := greeting.Greet(*name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(text)
}
