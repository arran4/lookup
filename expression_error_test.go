package lookup

import (
	"errors"
	"testing"
)

func TestEvaluatorErrorIntegrity(t *testing.T) {
	errRunner := Error(errors.New("intentional error"))

	tests := []struct {
		name      string
		runner    Runner
		expectErr bool
	}{
		// Comparison error propagation
		{"GreaterThan with error runner", GreaterThan(errRunner), true},
		{"LessThan with error runner", LessThan(errRunner), true},
		{"GreaterThanOrEqual with error runner", GreaterThanOrEqual(errRunner), true},
		{"LessThanOrEqual with error runner", LessThanOrEqual(errRunner), true},
		{"NotEquals with error runner", NotEquals(errRunner), true},
		{"Equals with error runner", Equals(errRunner), true},

		// Boolean helpers error propagation
		{"Truthy with error runner", Truthy(errRunner), true},
		{"ToBool with error runner", ToBool(errRunner), true},
		{"Not with error runner", Not(errRunner), true},

		// Match
		{"Match with error runner", Match(errRunner), true},
	}

	scope := NewScope(nil, Constant(5)) // Base value is 5 for comparison

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.runner.Run(scope)
			_, isInvalidor := res.(*Invalidor)

			if tt.expectErr && !isInvalidor {
				t.Errorf("Expected an Invalidor error, but got %T (value: %v)", res, res.Raw())
			} else if !tt.expectErr && isInvalidor {
				t.Errorf("Did not expect an error, but got %v", res.Raw())
			}
		})
	}
}

func TestEvaluatorComparisonErrorIntegrity(t *testing.T) {
	errRunner := Error(errors.New("intentional error"))

	tests := []struct {
		name      string
		runner    Runner
		expectErr bool
	}{
		{"BinaryEquals left error", BinaryEquals(errRunner, Constant(1)), true},
		{"BinaryEquals right error", BinaryEquals(Constant(1), errRunner), true},
		{"BinaryNotEquals left error", BinaryNotEquals(errRunner, Constant(1)), true},
		{"BinaryNotEquals right error", BinaryNotEquals(Constant(1), errRunner), true},
		{"BinaryGreaterThan left error", BinaryGreaterThan(errRunner, Constant(1)), true},
		{"BinaryGreaterThan right error", BinaryGreaterThan(Constant(1), errRunner), true},
		{"BinaryLessThan left error", BinaryLessThan(errRunner, Constant(1)), true},
		{"BinaryLessThan right error", BinaryLessThan(Constant(1), errRunner), true},
		{"BinaryGreaterThanOrEqual left error", BinaryGreaterThanOrEqual(errRunner, Constant(1)), true},
		{"BinaryGreaterThanOrEqual right error", BinaryGreaterThanOrEqual(Constant(1), errRunner), true},
		{"BinaryLessThanOrEqual left error", BinaryLessThanOrEqual(errRunner, Constant(1)), true},
		{"BinaryLessThanOrEqual right error", BinaryLessThanOrEqual(Constant(1), errRunner), true},
	}

	scope := NewScope(nil, Constant(5)) // Base value is 5 for comparison

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.runner.Run(scope)
			_, isInvalidor := res.(*Invalidor)

			if tt.expectErr && !isInvalidor {
				t.Errorf("Expected an Invalidor error, but got %T (value: %v)", res, res.Raw())
			} else if !tt.expectErr && isInvalidor {
				t.Errorf("Did not expect an error, but got %v", res.Raw())
			}
		})
	}
}

func TestEvaluatorIfErrorIntegrity(t *testing.T) {
	errRunner := Error(errors.New("intentional error"))

	tests := []struct {
		name      string
		runner    Runner
		expectErr bool
	}{
		{"If condition error", If(errRunner, Constant(1), Constant(2)), true},
		// NOTE: In the evaluator library, BoolType is forgiving and treats non-nil objects (like an array) as true.
		// There is no explicit "type mismatch error" generated from Evaluate() for a boolean truthiness check on an array.
		{"If condition type mismatch", If(Constant([]int{1, 2, 3}), Constant(1), Constant(2)), false},
	}

	scope := NewScope(nil, Constant(5)) // Base value is 5 for comparison

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.runner.Run(scope)
			_, isInvalidor := res.(*Invalidor)

			if tt.expectErr && !isInvalidor {
				t.Errorf("Expected an Invalidor error, but got %T (value: %v)", res, res.Raw())
			} else if !tt.expectErr && isInvalidor {
				t.Errorf("Did not expect an error, but got %v", res.Raw())
			}
		})
	}
}
