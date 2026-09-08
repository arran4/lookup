package lookup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type strictTestNode struct {
	A struct {
		B []struct {
			C int
		}
	}
	M map[string]interface{}
}

func newStrictTestNode() *strictTestNode {
	n := &strictTestNode{
		M: map[string]interface{}{
			"A.B": []interface{}{
				map[string]interface{}{"C": 3},
			},
			"a!@#$%^&*()_+{}|:<>?~": map[string]interface{}{
				"B": 4,
			},
			"A B":  5,
			"日本語":  map[string]interface{}{"C": 6},
			"😊👍":   []interface{}{7},
			"A\"B": 8,
		},
	}
	n.A.B = []struct{ C int }{{C: 1}, {C: 2}}
	return n
}

func TestCompileSimplePath(t *testing.T) {
	cases := []struct {
		name  string
		query string
		valid bool
		want  interface{}
	}{
		// Valid cases that must resolve
		{"empty", "", true, newStrictTestNode()},
		{"basic", "A.B[0].C", true, 1},
		{"leading dot", ".A.B[0].C", true, 1},
		{"quoted basic", `M."A.B"[0].C`, true, 3},
		{"quoted punctuation", `M."a!@#$%^&*()_+{}|:<>?~".B`, true, 4},
		{"escaped quote", `M."A\"B"`, true, 8},
		{"backslash escape", `M.A\ B`, true, 5},
		{"unicode basic", `M.日本語.C`, true, 6},
		{"unicode quote", `M."😊👍"[0]`, true, 7},

		// Invalid cases that must fail compilation
		{"invalid char after bracket", `A[0]B`, false, nil},
		{"invalid char after bracket quote", `A[0]"B"`, false, nil},
		{"empty quoted key", `A.""`, false, nil},
		{"empty root quote", `""`, false, nil},
		{"bracket after dot", `A.[0]`, false, nil},
		{"bracket string key", `A["B"]`, false, nil},
		{"bracket non numeric", `A[B]`, false, nil},
		{"bracket escape bracket", `A[\[0\]]`, false, nil},
		{"trailing dot", "A.", false, nil},
		{"empty bracket", "A[]", false, nil},
		{"empty root", ".", false, nil},
		{"unmatched bracket", "A.B[0", false, nil},
		{"unmatched quote", `A."B`, false, nil},
		{"unexpected char after quote", `"A"B`, false, nil},
		{"unexpected quote", `A"B"`, false, nil},
		{"empty key", "A..B", false, nil},
		{"trailing escape", `A\`, false, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rel, err := CompileSimplePath(c.query)
			if c.valid {
				assert.NoError(t, err)
				if c.want != nil && err == nil {
					res := rel.Run(NewScope(nil, Reflect(newStrictTestNode())))
					assert.Equal(t, c.want, res.Raw())
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}
