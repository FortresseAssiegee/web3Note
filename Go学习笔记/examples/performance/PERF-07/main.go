// 运行示例：PERF-07
package main

import (
	"bytes"
	"fmt"
	"runtime/pprof"
)

//go:noinline
func waitForRelease(ready chan<- struct{}, release <-chan struct{}, done chan<- struct{}) {
	close(ready)
	<-release
	close(done)
}

func main() {
	ready := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	go waitForRelease(ready, release, done)
	<-ready
	var stacks bytes.Buffer
	if err := pprof.Lookup("goroutine").WriteTo(&stacks, 2); err != nil {
		panic(err)
	}
	fmt.Println("worker visible:", bytes.Contains(stacks.Bytes(), []byte("main.waitForRelease")))
	close(release)
	<-done
	fmt.Println("worker joined")
}
