package lookup

import (
	"testing"
)

func TestElementOfRegression(t *testing.T) {
	type MyInt int

	// Named type strictness
	t.Run("NamedIntStrictness", func(t *testing.T) {
		slice := []int{1, 2, 3}
		val := MyInt(1)

		// Should be false because types differ
		scope := NewScope(nil, Reflect(slice))
		r := Contains(Constant(val)).Run(scope)
		if r.Raw().(bool) {
			t.Errorf("Expected false for MyInt(1) in []int{1}, got true")
		}
	})

	// Interface slice with named types
	t.Run("InterfaceNamedStrictness", func(t *testing.T) {
		slice := []interface{}{1, 2, 3}
		val := MyInt(1)

		scope := NewScope(nil, Reflect(slice))
		r := Contains(Constant(val)).Run(scope)
		if r.Raw().(bool) {
			t.Errorf("Expected false for MyInt(1) in []interface{}{1}, got true")
		}
	})
}
