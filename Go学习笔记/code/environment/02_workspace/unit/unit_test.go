package unit

import "testing"

func TestKiB(t *testing.T) {
	got, err := KiB(2048)
	if err != nil || got != "2 KiB" {
		t.Fatalf("got %q, %v", got, err)
	}
	for _, input := range []int64{-1024, 1025} {
		if _, err := KiB(input); err == nil {
			t.Fatalf("accepted invalid input %d", input)
		}
	}
}
