package jsonata

import (
	"testing"

	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
)

func TestCompileBinaryOperators(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		left     interface{}
		right    interface{}
		expected interface{}
	}{
		{"Equals True", "=", 1, 1, true},
		{"Equals False", "=", 1, 2, false},
		{"NotEquals True", "!=", 1, 2, true},
		{"NotEquals False", "!=", 1, 1, false},
		{"GreaterThan True", ">", 2, 1, true},
		{"GreaterThan False", ">", 1, 2, false},
		{"LessThan True", "<", 1, 2, true},
		{"LessThan False", "<", 2, 1, false},
		{"GreaterThanOrEqual True", ">=", 1, 1, true},
		{"GreaterThanOrEqual False", ">=", 1, 2, false},
		{"LessThanOrEqual True", "<=", 1, 1, true},
		{"LessThanOrEqual False", "<=", 2, 1, false},
		{"In True", "in", 1, []interface{}{1, 2}, true},
		{"In False", "in", 3, []interface{}{1, 2}, false},
		{"And True", "and", true, true, true},
		{"And False", "and", true, false, false}, // Returns false
		{"And Return RHS", "and", true, 5, 5},
		{"And Return LHS", "and", false, 5, false},
		{"Or True", "or", true, false, true},
		{"Or Return LHS", "or", 5, false, 5},
		{"Or Return RHS", "or", false, 5, 5},
		{"Sequence", "..", 1, 3, []interface{}{int64(1), int64(2), int64(3)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ast := &AST{
				Node: &BinaryNode{
					Operator: test.operator,
					Left:     &LiteralNode{Value: test.left},
					Right:    &LiteralNode{Value: test.right},
				},
			}
			runner := Compile(ast)
			// Empty scope
			scope := lookup.NewScope(lookup.NewConstantor("", nil), lookup.NewConstantor("", nil))
			res := runner.Run(scope)

			if expectedSlice, ok := test.expected.([]interface{}); ok {
				// Handle slice comparison
				// Sequence creates []interface{}, simple elements match is usually enough
				// But we should check exact order for sequence usually.
				// Assert Equal should work for slices of same type.
				assert.Equal(t, expectedSlice, res.Raw())
			} else {
				assert.Equal(t, test.expected, res.Raw())
			}
		})
	}
}

func TestCompileUnsupportedBinary(t *testing.T) {
	ast := &AST{
		Node: &BinaryNode{
			Operator: "???",
			Left:     &LiteralNode{Value: 1},
			Right:    &LiteralNode{Value: 1},
		},
	}
	runner := Compile(ast)
	scope := lookup.NewScope(lookup.NewConstantor("", nil), lookup.NewConstantor("", nil))
	res := runner.Run(scope)

	_, ok := res.(*lookup.Invalidor)
	assert.True(t, ok, "Expected Invalidor for unsupported operator")
}
