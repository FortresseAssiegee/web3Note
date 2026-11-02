package workload

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

func FormatGeneric(symbol string, count int) string {
	return fmt.Sprintf("%s:%d", symbol, count)
}

func FormatSpecific(symbol string, count int) string {
	var b strings.Builder
	b.Grow(len(symbol) + 24)
	b.WriteString(symbol)
	b.WriteByte(':')
	b.WriteString(strconv.Itoa(count))
	return b.String()
}

func Checksum(iterations int) uint64 {
	value := uint64(1)
	for i := 0; i < iterations; i++ {
		value ^= uint64(i) + 0x9e3779b97f4a7c15
		value = (value << 7) | (value >> 57)
		value *= 0xbf58476d1ce4e5b9
	}
	return value
}

// Allocate retains one buffer in every keepEvery allocations.
// All arguments must be positive; callers validate them before use.
func Allocate(count, size, keepEvery int) [][]byte {
	var kept [][]byte
	for i := 0; i < count; i++ {
		buffer := make([]byte, size)
		buffer[0] = byte(i)
		if i%keepEvery == 0 {
			kept = append(kept, buffer)
		}
	}
	return kept
}

// Contended deliberately keeps simulated work inside a mutex.
// It models a problematic critical section, not recommended service code.
func Contended(workers, rounds int, pause time.Duration) int {
	var mu sync.Mutex
	var wg sync.WaitGroup
	total := 0
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < rounds; i++ {
				mu.Lock()
				if pause > 0 {
					time.Sleep(pause)
				}
				total++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return total
}
