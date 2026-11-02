// 运行示例：PERF-04
package main

import (
	"bytes"
	"fmt"
	"runtime/pprof"
)

var sink uint64

func main() {
	var recording bytes.Buffer
	if err := pprof.StartCPUProfile(&recording); err != nil {
		panic(err)
	}
	var sum uint64
	for i := uint64(0); i < 100000; i++ {
		sum += i * i
	}
	sink = sum
	pprof.StopCPUProfile()
	fmt.Println("computed:", sink > 0)
	fmt.Println("profile written:", recording.Len() > 0)
}
