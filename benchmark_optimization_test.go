package lookup

import (
	"reflect"
	"testing"
)

func BenchmarkArrayOrSliceForEachPath(b *testing.B) {
	size := 1000
	data := make([]int, size)
	for i := 0; i < size; i++ {
		data[i] = i
	}
	v := reflect.ValueOf(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arrayOrSliceForEachPath("root", nil, v, nil, nil)
	}
}
