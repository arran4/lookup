package lookup

import (
	"testing"
)

func BenchmarkIntersection(b *testing.B) {
	size := 1000
	left := make([]int, size)
	right := make([]int, size)
	for i := 0; i < size; i++ {
		left[i] = i
		right[i] = i + size/2
	}
	// Overlap will be from size/2 to size-1 (500 elements)

	r := Reflect(left)
	rightConstant := NewConstantor("", right)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Find("", Intersection(rightConstant))
	}
}
