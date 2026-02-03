package jsonata

import (
	"testing"

	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
)

func TestBinaryOperators(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected interface{}
	}{
		{"Multiply", "2 * 3", 6.0},
		{"Divide", "6 / 2", 3.0},
		{"Modulo", "10 % 3", 1},
		{"Subtract", "5 - 2", 3.0},
		{"Equals", "5 = 5", true},
		{"NotEquals", "5 != 3", true},
		{"GreaterThan", "5 > 3", true},
		{"LessThan", "3 < 5", true},
		{"GreaterThanOrEqual", "5 >= 5", true},
		{"LessThanOrEqual", "3 <= 5", true},
		{"In", "5 in [1, 2, 5]", true},
		{"And", "true and true", true},
		{"Or", "false or true", true},
		{"Precedence Mult Add", "1 + 2 * 3", 7.0},
		{"Precedence Add Compare", "1 + 1 = 2", true},
		{"And Number", "1 and 2", true},
		{"Or Number", "0 or 1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			r := Compile(ast)
			// Empty data
			rootVal := lookup.Reflect(map[string]interface{}{})
			root := lookup.NewScope(rootVal, rootVal)
			res := r.Run(root)
			got := res.Raw()

			if assert.NotNil(t, got) {
				// Handle numeric type differences (int vs float)
				if val, ok := got.(int); ok {
					got = float64(val)
				} else if val, ok := got.(int64); ok {
					got = float64(val)
				}

				expected := tt.expected
				if val, ok := expected.(int); ok {
					expected = float64(val)
				}

				assert.Equal(t, expected, got)
			}
		})
	}
}
