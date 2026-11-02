// 运行示例：ENV-02
package main

import (
	"fmt"
	"os"
)

func main() {
	const key = "GO_NOTES_ENV_DEMO"
	old, existed := os.LookupEnv(key)
	defer func() {
		if existed {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	}()
	if err := os.Unsetenv(key); err != nil {
		panic(err)
	}
	value, ok := os.LookupEnv(key)
	fmt.Printf("missing: value=%q present=%v\n", value, ok)
	if err := os.Setenv(key, ""); err != nil {
		panic(err)
	}
	value, ok = os.LookupEnv(key)
	fmt.Printf("empty: value=%q present=%v\n", value, ok)
}
