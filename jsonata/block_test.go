package jsonata

import (
	"reflect"
	"testing"

	"github.com/arran4/lookup"
)

func TestBlockSemantics(t *testing.T) {
	tests := []struct {
		name          string
		expr          string
		expected      interface{}
		expectErr     bool
		expectedErrIs error
	}{
		{
			name:     "Empty block returns undefined",
			expr:     "()",
			expected: Undefined{},
		},
		{
			name:     "Block evaluates sequentially and returns last",
			expr:     "(1; 2; 3)",
			expected: float64(3),
		},
		{
			name:     "Trailing semicolon",
			expr:     "(1; 2; 3;)",
			expected: float64(3),
		},
		{
			name:     "Preservation of ordinary grouping",
			expr:     "(1 + 2)",
			expected: float64(3),
		},
		{
			name:      "Genuine error propagation",
			expr:      "(1; 1/0; 3)",
			expectErr: true,
			// Just check expectErr for now
		},
		{
			name:     "Earlier undefined doesn't prevent later execution",
			expr:     "(nonexistent; 42)",
			expected: float64(42),
		},
		{
			name:     "Postfix composition behavior preservation",
			expr:     "([1, 2, 3])[1]",
			expected: float64(2),
		},
		{
			name:     "Postfix composition on block",
			expr:     "(1; 2; [3, 4])[1]",
			expected: float64(4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := Parse(tc.expr)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			runner := Compile(ast)
			scope := lookup.NewScope(nil, nil)
			res := runner.Run(scope)

			if inv, ok := res.(*lookup.Invalidor); ok {
				if !tc.expectErr {
					t.Fatalf("Unexpected execution error: %v", inv.Unwrap())
				}

				return
			}

			if tc.expectErr {
				t.Fatalf("Expected execution error but got none")
			}

			var val interface{}
			if res != nil {
				val = res.Raw()
			}
			val = Materialize(val)

			if !reflect.DeepEqual(val, tc.expected) {
				t.Errorf("Expected %#v, got %#v", tc.expected, val)
			}
		})
	}
}
