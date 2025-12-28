package lookup

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestSimpleor(t *testing.T) {
	data := map[string]interface{}{
		"foo": "bar",
		"nested": map[string]interface{}{
			"baz": 123,
		},
		"list": []interface{}{1, 2, 3},
	}

	s := Simple(data)

	tests := []struct {
		name string
		path string
		want interface{}
		fail bool
	}{
		{name: "root", path: "", want: data},
		{name: "map lookup", path: "foo", want: "bar"},
		{name: "nested map lookup", path: "nested", want: map[string]interface{}{"baz": 123}},
		{name: "list lookup", path: "list", want: []interface{}{1, 2, 3}},
		{name: "missing key", path: "missing", fail: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Find(tt.path)
			if tt.fail {
				if _, ok := got.(*Invalidor); !ok {
					t.Errorf("expected failure, got %T: %v", got, got.Raw())
				}
				return
			}
			if diff := cmp.Diff(tt.want, got.Raw()); diff != "" {
				t.Errorf("unexpected result: %s", diff)
			}
		})
	}
}

func TestSimpleorModifiers(t *testing.T) {
	data := map[string]interface{}{
		"list": []interface{}{1, 2, 3},
	}
	s := Simple(data)

	// Test Index modifier with Simpleor
	got := s.Find("list", Index(1))
	if diff := cmp.Diff(2, got.Raw()); diff != "" {
		t.Errorf("Index(1) unexpected result: %s", diff)
	}
}

func TestSimpleorFallbackPath(t *testing.T) {
	type Struct struct{ Name string }
	data := map[string]interface{}{
		"s": Struct{Name: "foo"},
	}
	s := Simple(data)

	// This should trigger fallback for "s" and then find "Name" via reflection
	// The path should be "s.Name"
	r := s.Find("s").Find("Name")

	if diff := cmp.Diff("foo", r.Raw()); diff != "" {
		t.Errorf("Value mismatch: %s", diff)
	}

	// Path check depends on how PathBuilder builds it.
	// Simpleor.Find("s") -> path "s"
	// Reflector("s").Find("Name") -> path "s.Name" (if initialized with path "s")
	if p := ExtractPath(r); p != "s.Name" {
		t.Errorf("Path mismatch: expected 's.Name', got '%s'", p)
	}
}
