package greeting_test

import (
	"errors"
	"example.com/learning/hello/internal/greeting"
	"fmt"
	"testing"
)

func TestGreet(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"plain", "Gopher", "Hello, Gopher!"},
		{"trim", "  Go  ", "Hello, Go!"},
		{"unicode", "学习者", "Hello, 学习者!"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := greeting.Greet(tc.input)
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestEmptyName(t *testing.T) {
	_, err := greeting.Greet(" \t\n")
	if !errors.Is(err, greeting.ErrEmptyName) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func ExampleGreet() {
	text, err := greeting.Greet("Gopher")
	fmt.Println(text, err)
	// Output: Hello, Gopher! <nil>
}

func BenchmarkGreet(b *testing.B) {
	for b.Loop() {
		if _, err := greeting.Greet("Gopher"); err != nil {
			b.Fatal(err)
		}
	}
}
