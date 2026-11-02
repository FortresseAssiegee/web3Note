// 运行示例：PRJ-06
package main

import (
	"fmt"
	"io"
	"strings"
)

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

func read(body io.ReadCloser, limit int64) error {
	defer body.Close()
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > limit {
		return fmt.Errorf("response too large")
	}
	return nil
}

func main() {
	small := &trackedBody{Reader: strings.NewReader("ok")}
	large := &trackedBody{Reader: strings.NewReader("toolong")}
	fmt.Println(read(small, 3) == nil, small.closed)
	fmt.Println(read(large, 3) != nil, large.closed)
}
