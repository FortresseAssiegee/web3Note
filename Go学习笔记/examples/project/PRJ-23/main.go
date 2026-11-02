// 运行示例：PRJ-23
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
)

func main() {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	slog.New(handler).Info("quote_refreshed", "symbol", "BTCUSDT", "source", "demo")
	var fields map[string]any
	if err := json.Unmarshal(buf.Bytes(), &fields); err != nil {
		panic(err)
	}
	fmt.Println(fields["level"], fields["msg"], fields["symbol"], fields["source"])
}
