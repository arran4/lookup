package lookup

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestIf(t *testing.T) {
	type data struct{ Value bool }

	trueObj := &data{Value: true}
	falseObj := &data{Value: false}

	tests := []struct {
		name      string
		obj       *data
		cond      Runner
		then      Runner
		otherwise Runner
		want      interface{}
		fail      bool
	}{
		{name: "constant true", obj: trueObj, cond: Constant(true), then: Constant("yes"), otherwise: Constant("no"), want: "yes"},
		{name: "constant false", obj: trueObj, cond: Constant(false), then: Constant("yes"), otherwise: Constant("no"), want: "no"},
		{name: "scope true", obj: trueObj, cond: ToBool(This()), then: Constant("yes"), otherwise: Constant("no"), want: "yes"},
		{name: "scope false", obj: falseObj, cond: ToBool(This()), then: Constant("yes"), otherwise: Constant("no"), want: "no"},
		{name: "invalid condition", obj: trueObj, cond: Constant("bad"), then: Constant("yes"), otherwise: Constant("no"), fail: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Reflect(tt.obj).Find("Value", If(tt.cond, tt.then, tt.otherwise))
			if tt.fail {
				if _, ok := r.(*Invalidor); !ok {
					t.Errorf("expected invalid result, got %#v", r.Raw())
				}
				return
			}
			if diff := cmp.Diff(tt.want, r.Raw()); diff != "" {
				t.Errorf("If() mismatch: %s", diff)
			}
		})
	}
}
