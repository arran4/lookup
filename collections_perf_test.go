package lookup

import (
	"testing"
)

func BenchmarkFirst(b *testing.B) {
	size := 10000
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = i
	}
	// We want to find the last element to maximize iterations
	target := size - 1
	r := Reflect(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use First to find the element that equals target
		// This forces iteration through the whole slice constructing paths
		r.Find("", First(Equals(Constant(target)))).Raw()
	}
}

func BenchmarkLast(b *testing.B) {
	size := 10000
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = i
	}
	// We want to find the first element (0) using Last to maximize iterations
	// Last iterates backwards
	target := 0
	r := Reflect(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Find("", Last(Equals(Constant(target)))).Raw()
	}
}

func BenchmarkForEach(b *testing.B) {
	// forEach is internal but we can trigger it via Map or Filter
	// Filter uses arrayOrSliceForEachPath which calls forEach logic.
	// So if we Filter a large array, it will iterate and construct paths.

	size := 10000
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = i
	}
	r := Reflect(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Filter everything, so it processes all elements
		r.Find("", Filter(GreaterThan(Constant(-1)))).Raw()
	}
}
