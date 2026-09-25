package jsonata

import (
	"encoding/json"
	"testing"

	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestNode struct {
	Name     string
	Size     int
	Tags     []string
	Children []*TestNode
}

func runQuery(t *testing.T, data interface{}, q string) interface{} {
	ast, err := Parse(q)
	assert.NoError(t, err)
	r := Compile(ast)
	root := lookup.Reflect(data)
	res := r.Run(lookup.NewScope(root, root))
	return res.Raw()
}

func TestStructQueries(t *testing.T) {
	root := &TestNode{
		Name: "root",
		Size: 3,
		Children: []*TestNode{
			{Name: "child1", Size: 1},
			{Name: "child2", Size: 2},
		},
	}

	assert.Equal(t, "root", runQuery(t, root, "Name"))
	assert.Equal(t, "child2", runQuery(t, root, "Children[1].Name"))
	assert.Equal(t, 2, runQuery(t, root, "Children[Name='child2'].Size"))
}

func TestJSONQueries(t *testing.T) {
	jsonData := []byte(`{"users":[{"name":"bob","age":5},{"name":"sam","age":7}]}`)
	var v struct {
		Users []struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"users"`
	}
	if err := json.Unmarshal(jsonData, &v); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	assert.Equal(t, 7, runQuery(t, v, "Users[Name='sam'].Age"))
}

func TestArrayConstructorEvaluation(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		input   interface{}
		want    interface{}
		wantErr bool
	}{
		{"empty array", `[]`, nil, []interface{}{}, false},
		{"singleton array", `[1]`, nil, []interface{}{1.0}, false}, // JSON numbers usually test as float64
		{"literal array", `[1, 2]`, nil, []interface{}{1.0, 2.0}, false},
		{"nested literal array", `[[1, 2], 3]`, nil, []interface{}{[]interface{}{1.0, 2.0}, 3.0}, false},
		{"arithmetic expression", `[1 + 2, 4 * 2]`, nil, []interface{}{3.0, 8.0}, false},
		{"field paths evaluated against input", `[foo, bar]`, map[string]interface{}{"foo": 1, "bar": 2}, []interface{}{1, 2}, false},
		{"nested expression valued arrays", `[[foo], bar]`, map[string]interface{}{"foo": 1, "bar": 2}, []interface{}{[]interface{}{1}, 2}, false},
		{"missing field vs explicit null", `[foo, bar]`, map[string]interface{}{"foo": nil}, []interface{}{nil}, false}, // bar is undefined/missing and therefore omitted
		{"context is not mutated regression", `[foo, foo]`, map[string]interface{}{"foo": 1}, []interface{}{1, 1}, false},
		{"failing element expression", `[1/0]`, nil, nil, true}, // division by zero

		// PR Review Requirements
		{"absent result omission (nil Pathor)", `[nonexistent]`, nil, []interface{}{}, false},                     // missing from un-keyed input -> nil Pathor -> omitted
		{"explicit null preservation", `[null]`, nil, []interface{}{nil}, false},                                  // JSON null literal -> literal node -> evaluated to explicit null
		{"missing field omission (Invalidor)", `[foo]`, map[string]interface{}{"bar": 1}, []interface{}{}, false}, // specific map missing field Invalidor -> omitted
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			runner := Compile(ast)

			scope := lookup.NewScopeWithContext(lookup.Reflect(tt.input), lookup.Reflect(tt.input), nil)
			res := runner.Run(scope)

			if inv, ok := res.(*lookup.Invalidor); ok {
				if tt.wantErr {
					return
				}
				t.Fatalf("Evaluate error: %v", inv.Unwrap())
			}

			if tt.wantErr {
				t.Fatalf("Expected error, got nil")
			}

			raw := Materialize(res.Raw())

			require.Equal(t, tt.want, raw)
		})
	}
}

// jsonataArrayRunner unit tests directly testing inner pathor resolution behaviour
func TestJSONataArrayRunner_Unit(t *testing.T) {
	// 1. Omit nil / typed nil pathor
	// 2. Preserve explicit null
	// 3. Omit Invalidor

	tests := []struct {
		name string
		mock lookup.Runner
		want interface{}
	}{
		{"omit nil Pathor", &resultRunner{result: nil}, []interface{}{}},
		{"omit typed-nil Pathor", &resultRunner{result: (*lookup.Reflector)(nil)}, []interface{}{}},
		{"preserve explicit null", lookup.Constant(nil), []interface{}{nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &jsonataArrayRunner{elements: []lookup.Runner{tt.mock}}
			scope := lookup.NewScopeWithContext(lookup.Reflect(nil), lookup.Reflect(nil), nil)
			res := runner.Run(scope)

			require.NotNil(t, res)
			raw := Materialize(res.Raw())
			require.Equal(t, tt.want, raw)
		})
	}
}
