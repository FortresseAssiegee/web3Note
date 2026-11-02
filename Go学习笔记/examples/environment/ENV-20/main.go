// 运行示例：ENV-20
package main

import (
	"fmt"
	"strconv"
)

func choosePort(cli *string, envValue string, envPresent bool) (int, string, error) {
	raw, source := "8080", "default"
	if envPresent {
		raw, source = envValue, "env"
	}
	if cli != nil {
		raw, source = *cli, "cli"
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, source, fmt.Errorf("invalid port from %s", source)
	}
	return port, source, nil
}

func main() {
	cli := "7070"
	cases := []struct {
		name    string
		cli     *string
		env     string
		present bool
	}{
		{"default", nil, "", false},
		{"environment", nil, "9090", true},
		{"command", &cli, "9090", true},
		{"empty-env", nil, "", true},
	}
	for _, item := range cases {
		port, source, err := choosePort(item.cli, item.env, item.present)
		if err != nil {
			fmt.Printf("%s: rejected=true source=%s\n", item.name, source)
			continue
		}
		fmt.Printf("%s: port=%d source=%s\n", item.name, port, source)
	}
}
