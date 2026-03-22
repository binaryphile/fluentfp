package toc

import "testing"

func BenchmarkServiceTimeDistMemory(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = newHist()
	}
}
