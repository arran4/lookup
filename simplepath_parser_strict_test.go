package lookup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileSimplePath(t *testing.T) {
	cases := []struct {
		name    string
		query   string
		valid   bool
		runWith *testNode // Using testNode from simplepath_parser_test.go
		want    interface{}
	}{
		{"empty", "", true, nil, nil},
		{"basic", "A.B[0].C", true, newTestNode(), 1},
		{"leading dot", ".A.B[0].C", true, newTestNode(), 1},
		{"quoted basic", `A."B"[0].C`, true, newTestNode(), 1},
		{"quoted special", `"A.B"[0].C`, true, nil, nil},
		{"escaped quote", `"A\"B"`, true, nil, nil},
		{"backslash escape", `A\ B`, true, nil, nil},
		{"bracket index", `A[0]`, true, nil, nil},
		{"bracket string key", `A["B"]`, true, nil, nil},
		{"unicode basic", `A.日本語.C`, true, nil, nil},
		{"unicode quote", `"😊👍"[0]`, true, nil, nil},
		{"punctuation heavy", `"a!@#$%^&*()_+{}|:<>?~".B`, true, nil, nil},
		{"bracket escape bracket", `A[\[0\]]`, true, nil, nil},
		{"unmatched bracket", "A.B[0", false, nil, nil},
		{"unmatched quote", `A."B`, false, nil, nil},
		{"unexpected char after quote", `"A"B`, false, nil, nil},
		{"unexpected quote", `A"B"`, false, nil, nil},
		{"empty key", "A..B", false, nil, nil},
		{"trailing escape", `A\`, false, nil, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rel, err := CompileSimplePath(c.query)
			if c.valid {
				assert.NoError(t, err)
				if c.runWith != nil && err == nil {
					res := rel.Run(NewScope(nil, Reflect(c.runWith)))
					assert.Equal(t, c.want, res.Raw())
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}
