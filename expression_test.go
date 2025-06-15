package lookup

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestToBool_WithRunner(t *testing.T) {
	scope := NewScope(nil, NewConstantor("", false))
	got := ToBool(Constant("true")).Run(scope)
	if diff := cmp.Diff(true, got.Raw()); diff != "" {
		t.Fatalf("unexpected result: %v", diff)
	}
}

func TestTruthy_WithRunner(t *testing.T) {
	scope := NewScope(nil, NewConstantor("", false))
	got := Truthy(Constant(1)).Run(scope)
	if diff := cmp.Diff(true, got.Raw()); diff != "" {
		t.Fatalf("unexpected result: %v", diff)
	}
}
