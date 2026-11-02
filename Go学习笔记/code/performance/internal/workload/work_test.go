package workload

import "testing"

func TestFormatEquivalent(t *testing.T) {
	cases := []struct {
		symbol string
		count  int
	}{
		{"BTCUSDT", 12},
		{"", 0},
		{"中文", -7},
		{"A:B", 123456},
	}
	for _, tc := range cases {
		got := FormatSpecific(tc.symbol, tc.count)
		want := FormatGeneric(tc.symbol, tc.count)
		if got != want {
			t.Fatalf("symbol=%q count=%d: got %q, want %q", tc.symbol, tc.count, got, want)
		}
	}
}

func TestAllocateRetention(t *testing.T) {
	kept := Allocate(5, 16, 2)
	if len(kept) != 3 {
		t.Fatalf("kept %d buffers, want 3", len(kept))
	}
	for i, buffer := range kept {
		if len(buffer) != 16 || buffer[0] != byte(i*2) {
			t.Fatalf("invalid buffer %d: size=%d first=%d", i, len(buffer), buffer[0])
		}
	}
}

func TestContendedCompletesAllUpdates(t *testing.T) {
	if got := Contended(4, 25, 0); got != 100 {
		t.Fatalf("got %d, want 100", got)
	}
}

func TestChecksumInputMatters(t *testing.T) {
	if Checksum(0) != 1 {
		t.Fatal("empty workload should preserve seed")
	}
	if Checksum(10) == Checksum(11) {
		t.Fatal("these neighboring fixtures should differ")
	}
}
