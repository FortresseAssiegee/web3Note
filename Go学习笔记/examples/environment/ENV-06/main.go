// 运行示例：ENV-06
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Config struct {
	Port int `json:"port"`
}

func parseConfig(text string) (Config, error) {
	var cfg Config
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			err = errors.New("only one JSON object is allowed")
		}
		return cfg, err
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return cfg, errors.New("port is out of range")
	}
	return cfg, nil
}

func main() {
	cfg, err := parseConfig(`{"port":8080}`)
	if err != nil {
		panic(err)
	}
	fmt.Printf("port=%d\n", cfg.Port)
	_, err = parseConfig(`{"port":8080,"prot":9090}`)
	fmt.Printf("unknown rejected=%v\n", err != nil)
	_, err = parseConfig(`{"port":8080} {"port":9090}`)
	fmt.Printf("trailing rejected=%v\n", err != nil)
}
