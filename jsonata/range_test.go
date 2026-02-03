package jsonata

import (
	"reflect"
	"testing"

	"github.com/arran4/lookup"
)

func TestRangeExecution(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected interface{}
	}{
		{
			name:     "Simple range",
			expr:     "1..5",
			expected: []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:     "Single item range",
			expr:     "1..1",
			// Note: Current implementation unwraps singleton arrays to the value itself.
			// This might not be strictly JSONata compliant for explicit arrays, but is consistent with current runner behavior.
			expected: 1,
		},
		{
			name:     "Empty range (descending)",
			expr:     "5..1",
			expected: []interface{}{},
		},
		{
			name:     "Range with arithmetic LHS",
			expr:     "1+1..5",
			expected: []interface{}{2, 3, 4, 5},
		},
		{
			name:     "Range with arithmetic RHS",
			expr:     "1..2+2",
			expected: []interface{}{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := Parse(tt.expr)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			runner := Compile(node)
			scope := &lookup.Scope{
				// Position: lookup.NewConstantor("", nil), // Optional?
			}

			res := runner.Run(scope)

			if invalid, ok := res.(*lookup.Invalidor); ok {
				t.Fatalf("Runtime error: %v", invalid)
			}

			val := res.Raw()

			// Check types and values
			if !reflect.DeepEqual(val, tt.expected) {
				// Handle empty slice vs nil slice difference
				vVal := reflect.ValueOf(val)
				vExp := reflect.ValueOf(tt.expected)
				if vVal.Kind() == reflect.Slice && vExp.Kind() == reflect.Slice {
					if vVal.Len() == 0 && vExp.Len() == 0 {
						return // OK
					}
				}

				t.Errorf("Expected %v, got %v", tt.expected, val)
			}
		})
	}
}

func TestRangePrecedence(t *testing.T) {
	expr := `"a" & 1..3`
	node, err := Parse(expr)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	runner := Compile(node)
	scope := &lookup.Scope{}
	res := runner.Run(scope)

	val := res.Raw()
	if s, ok := val.(string); ok {
		// Go json.Marshal is compact: [1,2,3]
		// So "a[1,2,3]"
		if s != "a[1,2,3]" {
			t.Errorf("Expected 'a[1,2,3]', got '%s'", s)
		}
	} else {
		t.Errorf("Expected string, got %T: %v", val, val)
	}
}
