// 运行示例：PRJ-22
package main

import (
	"flag"
	"fmt"
	"io"
	"time"
)

func main() {
	fs := flag.NewFlagSet("quotes", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	addr := fs.String("addr", "127.0.0.1:8080", "listen address")
	ttl := fs.Duration("ttl", 5*time.Second, "cache TTL")
	if err := fs.Parse([]string{"-addr=127.0.0.1:9090", "-ttl=8s"}); err != nil {
		panic(err)
	}
	fmt.Println(*addr)
	fmt.Println(*ttl)
}
