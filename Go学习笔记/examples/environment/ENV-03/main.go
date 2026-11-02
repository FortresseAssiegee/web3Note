// 运行示例：ENV-03
package main

import (
	"flag"
	"fmt"
	"io"
)

func main() {
	flags := flag.NewFlagSet("preview", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	port := flags.Int("port", 8080, "server port")
	verbose := flags.Bool("verbose", false, "show details")
	if err := flags.Parse([]string{"-port", "9090", "-verbose", "input.txt"}); err != nil {
		panic(err)
	}
	fmt.Printf("port=%d verbose=%v args=%v\n", *port, *verbose, flags.Args())
}
