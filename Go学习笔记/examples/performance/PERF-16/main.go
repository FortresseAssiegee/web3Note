// 运行示例：PERF-16
package main

import (
	"fmt"
	"slices"
	"sync"
)

type Store struct {
	mu     sync.Mutex
	values map[string]int
}

func (s *Store) snapshot() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]int, len(s.values))
	for key, value := range s.values {
		result[key] = value
	}
	return result
}

func main() {
	store := Store{values: map[string]int{"ETH": 2, "BTC": 1}}
	snapshot := store.snapshot()
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		fmt.Println(key, snapshot[key])
	}
}
