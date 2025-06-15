package lookup

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

type ifNode struct {
	Name string
	Cond interface{}
}

func TestIf(t *testing.T) {
	tests := []struct {
		name          string
		node          *ifNode
		cond          Runner
		then          Runner
		otherwise     Runner
		expect        interface{}
		expectInvalid bool
	}{
		{
			name:      "true branch",
			node:      &ifNode{Name: "child", Cond: true},
			cond:      This("Cond"),
			then:      This("Name"),
			otherwise: Constant("other"),
			expect:    "child",
		},
		{
			name:      "false branch",
			node:      &ifNode{Name: "child", Cond: false},
			cond:      This("Cond"),
			then:      This("Name"),
			otherwise: Constant("other"),
			expect:    "other",
		},
		{
			name:      "string condition",
			node:      &ifNode{Name: "child", Cond: "true"},
			cond:      This("Cond"),
			then:      This("Name"),
			otherwise: Constant("other"),
			expect:    "child",
		},
		{
			name:      "int condition",
			node:      &ifNode{Name: "child", Cond: 0},
			cond:      This("Cond"),
			then:      This("Name"),
			otherwise: Constant("other"),
			expect:    "other",
		},
		{
			name:          "missing condition path",
			node:          &ifNode{Name: "child", Cond: true},
			cond:          This("Missing"),
			then:          This("Name"),
			otherwise:     Constant("other"),
			expectInvalid: true,
		},
		{
			name:          "unparsable type",
			node:          &ifNode{Name: "child", Cond: struct{}{}},
			cond:          This("Cond"),
			then:          This("Name"),
			otherwise:     Constant("other"),
			expectInvalid: true,
		},
		{
			name:      "nil then",
			node:      &ifNode{Name: "child", Cond: true},
			cond:      This("Cond"),
			then:      nil,
			otherwise: Constant("other"),
			expect:    &ifNode{Name: "child", Cond: true},
		},
		{
			name:      "nil otherwise",
			node:      &ifNode{Name: "child", Cond: false},
			cond:      This("Cond"),
			then:      This("Name"),
			otherwise: nil,
			expect:    &ifNode{Name: "child", Cond: false},
		},
		{
			name:          "nil value",
			node:          &ifNode{Name: "child", Cond: nil},
			cond:          This("Cond"),
			then:          This("Name"),
			otherwise:     Constant("other"),
			expectInvalid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Reflect(tt.node)
			res := r.Find("", If(tt.cond, tt.then, tt.otherwise))
			if tt.expectInvalid {
				if _, ok := res.(*Invalidor); !ok {
					t.Errorf("expected invalid result, got %v", res.Raw())
				}
				return
			}
			if diff := cmp.Diff(tt.expect, res.Raw()); diff != "" {
				t.Errorf("unexpected result: %s", diff)
			}
		})
	}
}
