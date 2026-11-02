// 运行示例：PERF-19
package main

import (
	"bytes"
	"fmt"
	"runtime"
	"runtime/pprof"
	"time"
)

func main() {
	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)
	done := make(chan struct{})
	timer := time.AfterFunc(5*time.Millisecond, func() { close(done) })
	<-done
	timer.Stop()
	var recording bytes.Buffer
	if err := pprof.Lookup("block").WriteTo(&recording, 0); err != nil {
		panic(err)
	}
	fmt.Println("block profile written:", recording.Len() > 0)
}
