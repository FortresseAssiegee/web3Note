// Package greeting creates a greeting after validating the name.
package greeting

import (
	"errors"
	"fmt"
	"strings"
)

var ErrEmptyName = errors.New("name must not be empty")

func Greet(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrEmptyName
	}
	return fmt.Sprintf("Hello, %s!", name), nil
}
