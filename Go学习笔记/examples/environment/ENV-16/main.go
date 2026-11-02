// 运行示例：ENV-16
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"testing/fstest"
)

func main() {
	_, err := fs.ReadFile(fstest.MapFS{}, "config.json")
	if err == nil {
		panic("expected missing file")
	}
	wrapped := fmt.Errorf("load application config: %w", err)
	var pathError *fs.PathError
	matched := errors.As(wrapped, &pathError)
	fmt.Printf("not-exist=%v\n", errors.Is(wrapped, fs.ErrNotExist))
	fmt.Printf("path-error=%v\n", matched)
	if matched {
		fmt.Printf("path=%s\n", pathError.Path)
	}
}
