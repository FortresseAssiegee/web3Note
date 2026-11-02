// Package unit formats exact nonnegative multiples of 1024 bytes.
package unit

import (
	"errors"
	"fmt"
)

func KiB(n int64) (string, error) {
	if n < 0 || n%1024 != 0 {
		return "", errors.New("need a nonnegative exact multiple of 1024")
	}
	return fmt.Sprintf("%d KiB", n/1024), nil
}
