package lookup

import (
	"reflect"
	"testing"
)

func TestForEachPath(t *testing.T) {
	data := []int{1, 2, 3}
	v := reflect.ValueOf(data)
	scope := NewScope(nil, &Reflector{path: "root", v: v})

	paths := []string{}
	err := forEach(scope, v, func(p Pathor) error {
		paths = append(paths, ExtractPath(p))
		return nil
	})

	if err != nil {
		t.Fatalf("forEach returned error: %v", err)
	}

	expected := []string{"root[0]", "root[1]", "root[2]"}
	if len(paths) != len(expected) {
		t.Fatalf("expected %d paths, got %d", len(expected), len(paths))
	}

	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("path at index %d: expected %s, got %s", i, expected[i], p)
		}
	}
}
