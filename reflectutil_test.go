package lookup

import (
	"errors"
	"reflect"
	"testing"
)

type structWithUnexported struct {
	Exported   int
	unexported int
}

func (s structWithUnexported) MethodVal() int {
	return s.Exported
}

func (s *structWithUnexported) MethodPtr() int {
	return s.Exported
}

func (s structWithUnexported) MethodErr(fail bool) (int, error) {
	if fail {
		return 0, errors.New("method error")
	}
	return s.Exported, nil
}

func (s structWithUnexported) MethodUnsupported(b bool) (int, bool) {
	return s.Exported, b
}

func (s structWithUnexported) MethodZero() {}

func (s structWithUnexported) MethodParams(x int) int {
	return x
}

func (s structWithUnexported) MethodThree() (int, int, int) {
	return 1, 2, 3
}

func TestElementOf(t *testing.T) {
	tests := []struct {
		name     string
		v        interface{}
		in       interface{}
		expected bool
	}{
		{
			name:     "ordinary slice - found",
			v:        2,
			in:       []int{1, 2, 3},
			expected: true,
		},
		{
			name:     "ordinary slice - not found",
			v:        4,
			in:       []int{1, 2, 3},
			expected: false,
		},
		{
			name:     "ordinary map - found",
			v:        "value2",
			in:       map[string]string{"key1": "value1", "key2": "value2"},
			expected: true,
		},
		{
			name:     "pointer-to-container - found",
			v:        2,
			in:       &[]int{1, 2, 3},
			expected: true,
		},
		{
			name:     "nil pointer",
			v:        2,
			in:       (*[]int)(nil),
			expected: false,
		},
		{
			name:     "zero-argument function returning a value/container",
			v:        2,
			in:       func() []int { return []int{1, 2, 3} },
			expected: true,
		},
		{
			name:     "zero-argument func() (T, error) with nil error",
			v:        2,
			in:       func() ([]int, error) { return []int{1, 2, 3}, nil },
			expected: true,
		},
		{
			name:     "func() (T, error) with non-nil error",
			v:        2,
			in:       func() ([]int, error) { return nil, errors.New("some error") },
			expected: false,
		},
		{
			name:     "unsupported func() (T, bool)",
			v:        2,
			in:       func() ([]int, bool) { return []int{1, 2, 3}, true },
			expected: false,
		},
		{
			name:     "zero-output functions",
			v:        2,
			in:       func() {},
			expected: false,
		},
		{
			name:     "functions/methods with parameters",
			v:        2,
			in:       func(x int) []int { return []int{x} },
			expected: false,
		},
		{
			name:     "more than two outputs",
			v:        2,
			in:       func() ([]int, int, int) { return []int{1, 2, 3}, 0, 0 },
			expected: false,
		},
		{
			name:     "struct with exported field",
			v:        2,
			in:       structWithUnexported{Exported: 2},
			expected: true,
		},
		{
			name:     "struct with unexported field ignored",
			v:        2,
			in:       structWithUnexported{unexported: 2},
			expected: false,
		},
		{
			name:     "struct method returning value",
			v:        2,
			in:       structWithUnexported{Exported: 2},
			expected: true,
		},
		{
			name:     "struct method err nil",
			v:        2,
			in:       func() (int, error) { return structWithUnexported{Exported: 2}.MethodErr(false) },
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic occurred: %v", r)
				}
			}()

			vVal := reflect.ValueOf(tt.v)
			inVal := reflect.ValueOf(tt.in)

			// We simulate what elementOf gets (we typically pass nil for pv initially)
			got := elementOf(vVal, inVal, nil)
			if got != tt.expected {
				t.Errorf("elementOf() = %v, want %v", got, tt.expected)
			}
		})
	}
}
