// 运行示例：PERF-27
package main

import (
	"bytes"
	"fmt"
	"sync"
)

func main() {
	pool := sync.Pool{New: func() any { return new(bytes.Buffer) }}
	buffer := pool.Get().(*bytes.Buffer)
	buffer.WriteString("Go")
	fmt.Println(buffer.String())
	buffer.Reset()
	pool.Put(buffer)

	next := pool.Get().(*bytes.Buffer)
	fmt.Println("empty:", next.Len() == 0)
	next.Reset()
	pool.Put(next)
}
