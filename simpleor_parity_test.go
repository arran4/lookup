package lookup_test

import (
	"github.com/arran4/lookup"
	"reflect"
	"testing"
)

func TestSimpleorReflectorParity(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
		"b": map[string]interface{}{
			"c": 2,
		},
		"list": []interface{}{
			"x", "y", "z",
		},
	}

	s := lookup.Simple(data)
	r := lookup.Reflect(data)

	tests := []struct {
		name string
		path string
		opts []lookup.Runner
	}{
		{"root", "", nil},
		{"nested_map", "b.c", nil},
		{"list_index", "list", []lookup.Runner{lookup.Index(1)}},
		{"this", "list", []lookup.Runner{lookup.This()}},
		{"parent", "list", []lookup.Runner{lookup.Index(1), lookup.Parent("")}},
		{"nested_scopes_chain", "list", []lookup.Runner{
			lookup.Chain(lookup.This(), lookup.Index(1)),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sRes := s.Find(tt.path, tt.opts...)
			rRes := r.Find(tt.path, tt.opts...)

			if !reflect.DeepEqual(sRes.Raw(), rRes.Raw()) {
				t.Errorf("Mismatch for path %q: Simpleor got %v, Reflector got %v", tt.path, sRes.Raw(), rRes.Raw())
			}
		})
	}
}

func TestSimpleorReflectorParity_ErrorPropagation(t *testing.T) {
	data := map[string]interface{}{}

	s := lookup.Simple(data)
	r := lookup.Reflect(data)

	sRes := s.Find("non_existent_key")
	rRes := r.Find("non_existent_key")

	// They both return Invalidor, but let's check equality semantics
	_, sIsInv := sRes.(*lookup.Invalidor)
	_, rIsInv := rRes.(*lookup.Invalidor)

	if sIsInv != rIsInv {
		t.Errorf("Mismatch in error propagation: Simpleor invalid=%v, Reflector invalid=%v", sIsInv, rIsInv)
	}
}

func TestSimpleorReflectorParity_NilValues(t *testing.T) {
	var data interface{} = nil

	s := lookup.Simple(data)
	r := lookup.Reflect(data)

	sRes := s.Find("any_key")
	rRes := r.Find("any_key")

	_, sIsInv := sRes.(*lookup.Invalidor)
	_, rIsInv := rRes.(*lookup.Invalidor)

	if sIsInv != rIsInv {
		t.Errorf("Mismatch for nil: Simpleor invalid=%v, Reflector invalid=%v", sIsInv, rIsInv)
	}
}

func TestSimpleorReflectorParity_TypedSlices(t *testing.T) {
	data := map[string]interface{}{
		"list": []int{1, 2, 3},
	}

	s := lookup.Simple(data)
	r := lookup.Reflect(data)

	sRes := s.Find("list")
	rRes := r.Find("list")

	if !reflect.DeepEqual(sRes.Raw(), rRes.Raw()) {
		t.Errorf("Mismatch for typed slices: Simpleor got %v, Reflector got %v", sRes.Raw(), rRes.Raw())
	}
}

func TestSimpleorReflectorParity_Result(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
	}

	s := lookup.Simple(data)
	r := lookup.Reflect(data)

	// Since we're dealing with context transitions in Scope, we should test lookup.Result (if it exists, though looking at code we only have This and Parent)
	// Let's test deep chaining with Relator which uses Scope traversal

	tests := []struct {
		name string
		path string
		opts []lookup.Runner
	}{
		{"parent", "a", []lookup.Runner{lookup.Parent("")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sRes := s.Find(tt.path, tt.opts...)
			rRes := r.Find(tt.path, tt.opts...)

			if !reflect.DeepEqual(sRes.Raw(), rRes.Raw()) {
				t.Errorf("Mismatch for path %q: Simpleor got %v, Reflector got %v", tt.path, sRes.Raw(), rRes.Raw())
			}
		})
	}
}
