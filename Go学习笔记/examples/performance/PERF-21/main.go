// 运行示例：PERF-21
package main

import (
	"bytes"
	"context"
	"fmt"
	"runtime/trace"
	"time"
)

func main() {
	recorder := trace.NewFlightRecorder(trace.FlightRecorderConfig{
		MinAge:   time.Second,
		MaxBytes: 1 << 20,
	})
	if err := recorder.Start(); err != nil {
		panic(err)
	}
	var total int
	trace.WithRegion(context.Background(), "work", func() {
		for i := 1; i <= 3; i++ {
			total += i
		}
	})
	var snapshot bytes.Buffer
	_, err := recorder.WriteTo(&snapshot)
	recorder.Stop()
	if err != nil {
		panic(err)
	}
	fmt.Println("result:", total)
	fmt.Println("snapshot written:", snapshot.Len() > 0)
}
