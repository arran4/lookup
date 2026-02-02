package lookup

import (
	"reflect"
	"testing"
)

func BenchmarkIntersectionHighCardinality(b *testing.B) {
	size := 1000
	left := make([]int, size)
	right := make([]int, size)
	for i := 0; i < size; i++ {
		left[i] = i
		right[i] = i + size/2 // Overlap by half
	}

	leftVal := reflect.ValueOf(left)
	rightVal := reflect.ValueOf(right)

	// Create a dummy scope
	scope := &Scope{
		Current: &Reflector{v: leftVal},
	}

	// Mock Runner for right side
	rightRunner := &mockRunner{
		val: rightVal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Re-create Intersection runner to reset state if any (it has none)
		op := Intersection(rightRunner)
		// Manually inject position since Intersection uses scope.Position
		scope.Position = &Reflector{v: leftVal}
		op.Run(scope)
	}
}

type mockRunner struct {
	val reflect.Value
}

func (m *mockRunner) Run(scope *Scope) Pathor {
	return &Reflector{v: m.val}
}

func (m *mockRunner) Raw() interface{} {
	return m.val.Interface()
}
