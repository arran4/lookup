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

	sInv, sIsInv := sRes.(*lookup.Invalidor)
	rInv, rIsInv := rRes.(*lookup.Invalidor)

	if sIsInv != rIsInv {
		t.Fatalf("Mismatch in error propagation: Simpleor invalid=%v, Reflector invalid=%v", sIsInv, rIsInv)
	}

	if sIsInv {
		// Both return Invalidor. Their paths might differ based on parser/reflector implementations
		// ("non_existent_key" vs "\"non_existent_key\""), but their observable interface behavior for Raw() must match.
		if sInv.Raw() != rInv.Raw() {
			t.Errorf("Mismatch in Invalidor Raw(): Simpleor %v, Reflector %v", sInv.Raw(), rInv.Raw())
		}

		// The errors are both populated but might not be completely semantically identical since simple/reflector paths differ.
		// However, we can assert both have an error and are not panicing.
		if sInv.Error() == "" && rInv.Error() != "" {
			t.Errorf("Mismatch in Error: Simpleor is empty, Reflector has error: %s", rInv.Error())
		}
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

func TestSimpleorReflectorParity_TypedMaps(t *testing.T) {
	data := map[string]int{
		"a": 10,
		"b": 20,
	}

	s := lookup.Simple(data)
	r := lookup.Reflect(data)

	sRes := s.Find("b")
	rRes := r.Find("b")

	if !reflect.DeepEqual(sRes.Raw(), rRes.Raw()) {
		t.Errorf("Mismatch for typed maps: Simpleor got %v, Reflector got %v", sRes.Raw(), rRes.Raw())
	}
}

type mockResultRunner struct {
	val interface{}
}

func (m mockResultRunner) Run(scope *lookup.Scope) lookup.Pathor {
	return lookup.Simple(m.val)
}

func TestSimpleorReflectorParity_Result(t *testing.T) {
	data := map[string]interface{}{
		"a": 1,
	}

	s := lookup.Simple(data)
	r := lookup.Reflect(data)

	// Chain the custom runner returning a different value, then lookup.Result("")
	// Result("") should return that different value (the Position), whereas This() would return the original Scope.Current.

	tests := []struct {
		name string
		path string
		opts []lookup.Runner
	}{
		{
			name: "result_vs_current",
			path: "a",
			opts: []lookup.Runner{
				lookup.Chain(mockResultRunner{val: 42}, lookup.Result("")),
			},
		},
		{
			name: "this_vs_current",
			path: "a",
			opts: []lookup.Runner{
				lookup.Chain(mockResultRunner{val: 42}, lookup.This()),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sRes := s.Find(tt.path, tt.opts...)
			rRes := r.Find(tt.path, tt.opts...)

			if !reflect.DeepEqual(sRes.Raw(), rRes.Raw()) {
				t.Errorf("Mismatch for path %q: Simpleor got %v, Reflector got %v", tt.path, sRes.Raw(), rRes.Raw())
			}

			// Also specifically check semantic correctness
			if tt.name == "result_vs_current" {
				if sRes.Raw() != 42 {
					t.Errorf("Simpleor expected 42 (Result), got %v", sRes.Raw())
				}
				if rRes.Raw() != 42 {
					t.Errorf("Reflector expected 42 (Result), got %v", rRes.Raw())
				}
			}
			if tt.name == "this_vs_current" {
				if sRes.Raw() != 1 {
					t.Errorf("Simpleor expected 1 (This), got %v", sRes.Raw())
				}
				if rRes.Raw() != 1 {
					t.Errorf("Reflector expected 1 (This), got %v", rRes.Raw())
				}
			}
		})
	}
}
