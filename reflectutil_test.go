package lookup

import (
	"errors"
	"reflect"
	"strings"
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

type structWithError struct {
	Value int
}

type customError struct {
	msg string
}

func (e customError) Error() string {
	return e.msg
}

func (s structWithError) MethodCustomErr(fail bool) (int, customError) {
	if fail {
		return 0, customError{msg: "failed"}
	}
	return s.Value, customError{}
}

type structWithMethod struct {
	Value int
}

func (s structWithMethod) ValueMethod() int {
	return s.Value * 10
}

func (s *structWithMethod) PointerMethod() int {
	return s.Value * 100
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
			v:        10, // Field is 2, method returns field * 5 = 10. Should match 10 via method, not field.
			in:       structWithMethod{Value: 1},
			expected: true,
		},
		{
			name:     "struct method err nil",
			v:        2,
			in:       func() (int, error) { return structWithUnexported{Exported: 2}.MethodErr(false) },
			expected: true,
		},
		{
			name:     "value-receiver method",
			v:        50,
			in:       structWithMethod{Value: 5},
			expected: true,
		},
		{
			name:     "pointer-receiver method",
			v:        500,
			in:       &structWithMethod{Value: 5},
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

func TestRunMethod(t *testing.T) {
	t.Run("rejects concrete custom error implementation", func(t *testing.T) {
		called := false
		f := func() (int, customError) {
			called = true
			return 1, customError{}
		}

		val := reflect.ValueOf(f)
		res := runMethod(val, "func")
		if res != nil {
			t.Errorf("Expected runMethod to return nil for unsupported signature with concrete custom error, got %v", res)
		}
		if called {
			t.Errorf("Expected runMethod to reject the function without calling it")
		}
	})

	t.Run("wraps underlying error context", func(t *testing.T) {
		sentinelErr := errors.New("my sentinel error")
		f := func() (int, error) {
			return 0, sentinelErr
		}

		val := reflect.ValueOf(f)
		res := runMethod(val, "fName")
		invalidor, ok := res.(*Invalidor)
		if !ok {
			t.Fatalf("Expected runMethod to return an Invalidor, got %T", res)
		}

		errResult := invalidor.err

		// Ensure context is kept
		expectedContext := "invalid element at simple path fName() method call returned error"
		if !strings.Contains(errResult.Error(), expectedContext) {
			t.Errorf("Expected error to contain context %q, but got %v", expectedContext, errResult)
		}

		// Ensure underlying error is preserved for errors.Is
		if !errors.Is(errResult, sentinelErr) {
			t.Errorf("Expected errors.Is to match sentinel error, but it did not. err = %v", errResult)
		}
	})
}
