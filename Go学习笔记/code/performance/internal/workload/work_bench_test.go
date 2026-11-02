package workload

import "testing"

var stringSink string
var numberSink uint64
var buffersSink [][]byte

func BenchmarkFormatGeneric(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		stringSink = FormatGeneric("BTCUSDT", 12)
	}
}

func BenchmarkFormatSpecific(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		stringSink = FormatSpecific("BTCUSDT", 12)
	}
}

func BenchmarkChecksum(b *testing.B) {
	for b.Loop() {
		numberSink = Checksum(1000)
	}
}

func BenchmarkAllocate(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		buffersSink = Allocate(100, 128, 16)
	}
}
