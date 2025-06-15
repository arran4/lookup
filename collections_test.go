package lookup

import (
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

func TestIndexConstantorPath(t *testing.T) {
	type Root struct{ Arr []int }
	root := &Root{Arr: []int{0, 1, 2}}

	r := Reflect(root).Find("Arr").Find("", Index(NewConstantor("Const", 1)))
	t.Logf("raw=%v path=%s", r.Raw(), ExtractPath(r))

	if diff := cmp.Diff(1, r.Raw()); diff != "" {
		t.Errorf("value mismatch: %s", diff)
	}
	if p := ExtractPath(r); !strings.HasPrefix(p, "Arr[1]") {
		t.Errorf("path mismatch got %s", p)
	}
}

func TestContainsMap(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	r := Reflect(m).Find("", Contains(Constant(2)))
	if diff := cmp.Diff(true, r.Raw()); diff != "" {
		t.Errorf("value mismatch: %s", diff)
	}
}

func TestInStruct(t *testing.T) {
	type S struct{ A, B string }
	s := S{A: "foo", B: "bar"}
	r := Reflect("bar").Find("", In(ValueOf(Reflect(s))))
	if diff := cmp.Diff(true, r.Raw()); diff != "" {
		t.Errorf("value mismatch: %s", diff)
	}
}
