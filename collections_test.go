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
