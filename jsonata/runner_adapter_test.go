package jsonata

import (
	"errors"
	"fmt"
	"github.com/arran4/go-evaluator"
	"github.com/arran4/lookup"
	"github.com/stretchr/testify/require"
	"testing"
)

type resultRunner struct {
	result lookup.Pathor
	calls  int
}

func (r *resultRunner) Run(_ *lookup.Scope) lookup.Pathor { r.calls++; return r.result }

func TestBinaryOperandResultsAndCounts(t *testing.T) {
	failure := lookup.NewInvalidor("", errors.New("real evaluator failure"))
	missing := lookup.NewInvalidor("", lookup.ErrNoSuchPath)
	var typedNil *lookup.Reflector
	for _, tc := range []struct {
		name, op    string
		left, right lookup.Pathor
		want        interface{}
		wantErr     *lookup.Invalidor
		rightCalls  int
	}{
		{"values once", "+", lookup.Reflect(2), lookup.Reflect(3), 5.0, nil, 1},
		{"nil left", "+", nil, lookup.Reflect(3), Undefined{}, nil, 1},
		{"nil right", "+", lookup.Reflect(2), nil, Undefined{}, nil, 1},
		{"typed nil", "+", typedNil, lookup.Reflect(3), Undefined{}, nil, 1},
		{"undefined", "+", lookup.Reflect(Undefined{}), lookup.Reflect(3), Undefined{}, nil, 1},
		{"missing left", "+", missing, lookup.Reflect(3), Undefined{}, nil, 1},
		{"missing right", "+", lookup.Reflect(3), missing, Undefined{}, nil, 1},
		{"nil concatenation", "&", nil, lookup.Reflect("x"), "x", nil, 1},
		{"nil comparison", "=", nil, lookup.Reflect(3), false, nil, 1},
		{"left error", "+", failure, lookup.Reflect(3), nil, failure, 0},
		{"right error", "+", lookup.Reflect(2), failure, nil, failure, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			left, right := &resultRunner{result: tc.left}, &resultRunner{result: tc.right}
			res := (&jsonataBinaryRunner{operator: tc.op, left: left, right: right}).Run(lookup.NewScope(nil, nil))
			require.Equal(t, 1, left.calls)
			require.Equal(t, tc.rightCalls, right.calls)
			if tc.wantErr != nil {
				require.Same(t, tc.wantErr, res)
			} else {
				require.Equal(t, tc.want, res.Raw())
			}
		})
	}
}

func TestSequenceRunnerResults(t *testing.T) {
	failure := lookup.NewInvalidor("", lookup.ErrEvalFail)
	for _, tc := range []struct {
		name   string
		result lookup.Pathor
		want   interface{}
	}{
		{"nil", nil, Undefined{}},
		{"undefined", lookup.Reflect(Undefined{}), Undefined{}},
		{"missing", lookup.NewInvalidor("", lookup.ErrNoSuchPath), Undefined{}},
		{"empty", lookup.Reflect([]int{}), &Sequence{Values: []interface{}{}}},
		{"typed range", lookup.Reflect([]int{1, 2}), &Sequence{Values: []interface{}{1, 2}}},
		{"null", lookup.Reflect(nil), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inner := &resultRunner{result: tc.result}
			res := (&jsonataSequenceRunner{inner: inner}).Run(lookup.NewScope(nil, nil))
			require.Equal(t, tc.want, res.Raw())
			require.Equal(t, 1, inner.calls)
		})
	}
	require.Same(t, failure, (&jsonataSequenceRunner{inner: &resultRunner{result: failure}}).Run(nil))
}

func TestPredicateNoMatchAndErrors(t *testing.T) {
	scope := lookup.NewScope(nil, lookup.Reflect([]int{1, 2}))
	for _, tc := range []struct {
		name   string
		result lookup.Pathor
		want   interface{}
	}{
		{"false", lookup.Reflect(false), Undefined{}},
		{"nil", nil, Undefined{}},
		{"missing", lookup.NewInvalidor("", lookup.ErrNoSuchPath), Undefined{}},
		{"true", lookup.Reflect(true), []interface{}{1, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pred := &resultRunner{result: tc.result}
			res := (&jsonataRunner{inner: &jsonataFilterRunner{predicate: pred}}).Run(scope)
			require.Equal(t, tc.want, res.Raw())
			require.Equal(t, 2, pred.calls)
		})
	}
	for _, err := range []error{errors.New("predicate failure"), lookup.ErrEvalFail, fmt.Errorf("inner: %w", lookup.ErrEvalFail), lookup.ErrNoMatchesForQuery} {
		failure := lookup.NewInvalidor("", err)
		pred := &resultRunner{result: failure}
		res := (&jsonataRunner{inner: &jsonataFilterRunner{predicate: pred}}).Run(scope)
		require.Same(t, failure, res)
		require.Equal(t, 1, pred.calls)
	}
	for _, expr := range []string{`rows[v = 9]`, `rows[missing = 1]`} {
		ast, err := Parse(expr)
		require.NoError(t, err)
		res := Compile(ast).Run(lookup.NewScope(nil, lookup.Reflect(map[string]interface{}{"rows": []map[string]interface{}{{"v": 1}, {"v": 2}}})))
		require.IsType(t, Undefined{}, res.Raw())
	}
}

func TestMaterializePrimitivesAndSequences(t *testing.T) {
	for _, value := range []interface{}{nil, false, 0, "", "hello", 42, int64(9), 3.5, true} {
		require.Equal(t, value, Materialize(value))
	}
	for _, tc := range []struct{ value, want interface{} }{
		{&Sequence{}, Undefined{}},
		{&Sequence{Values: []interface{}{nil}}, nil},
		{&Sequence{Values: []interface{}{1}}, 1},
		{&Sequence{Values: []interface{}{1, 2}}, []interface{}{1, 2}},
		{map[string]interface{}{"v": &Sequence{Values: []interface{}{&Array{Elements: []interface{}{1}}}}}, map[string]interface{}{"v": []interface{}{1}}},
	} {
		require.Equal(t, tc.want, Materialize(tc.value))
		res := (&jsonataRunner{inner: lookup.Constant(tc.value)}).Run(nil)
		require.Equal(t, tc.want, res.Raw())
	}
}

func TestScalarPathNoMatchAndGenuineErrors(t *testing.T) {
	for _, value := range []interface{}{1, "text", false, nil} {
		ast, err := Parse("missing")
		require.NoError(t, err)
		res := Compile(ast).Run(lookup.NewScope(nil, lookup.Reflect(value)))
		require.IsType(t, Undefined{}, res.Raw())
	}
	for _, err := range []error{errors.New("evaluator failure"), lookup.ErrEvalFail, lookup.ErrNoMatchesForQuery} {
		failure := lookup.NewInvalidor("", err)
		res := (&jsonataMapRunner{stepRunner: &resultRunner{result: failure}}).Run(lookup.NewScope(nil, lookup.Reflect(1)))
		require.Same(t, failure, res)
	}
}

type failingPathValue struct{}

var errPathMethod = errors.New("path method failed")

func (failingPathValue) V() (int, error) { return 0, errPathMethod }

func TestCompiledPredicateAndNestedPathErrors(t *testing.T) {
	for _, expr := range []string{"rows[V = 1]", "rows.V", "$sum(rows.V)"} {
		t.Run(expr, func(t *testing.T) {
			ast, err := Parse(expr)
			require.NoError(t, err)
			root := lookup.Reflect(map[string]interface{}{"rows": []interface{}{[]failingPathValue{{}}}})
			ctx := &evaluator.Context{Functions: GetStandardFunctions()}
			res := Compile(ast).Run(lookup.NewScopeWithContext(nil, root, ctx))
			inv, ok := res.(*lookup.Invalidor)
			require.True(t, ok, "result: %#v", res.Raw())
			require.ErrorIs(t, inv, errPathMethod)
		})
	}
}

func TestUndefinedClassificationRejectsErrorSubstrings(t *testing.T) {
	for _, message := range []string{
		"evaluator failed: element not found at simple path x element was map expected string",
		"evaluator failed: invalid element at simple path x element was int expected array,slice,map,struct,func",
		"element not found at simple path x element was map expected string: evaluator failure",
	} {
		inv := lookup.NewInvalidor("x", errors.New(message))
		require.False(t, IsUndefinedError(inv))
		require.False(t, isJSONataFieldNoMatch(inv))
		require.Same(t, inv, (&jsonataRunner{inner: &resultRunner{result: inv}}).Run(nil))
	}
}
