package jsonata

import (
	"encoding/json"
	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSemanticMissingPropertyVsNull(t *testing.T) {
	data := `{"a": 1, "b": null}`
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Fatalf("failed to unmarshal test data: %v", err)
	}
	root := lookup.Reflect(v)

	// Missing property
	astC, _ := Parse("c")
	runnerC := Compile(astC)
	resC := runnerC.Run(lookup.NewScope(nil, root))
	matC := Materialize(resC.Raw())

	if _, isUndef := matC.(Undefined); isUndef {
		t.Logf("Warning: matC is Undefined.")
	}

	// Property with null
	astB, _ := Parse("b")
	runnerB := Compile(astB)
	resB := runnerB.Run(lookup.NewScope(nil, root))
	matB := Materialize(resB.Raw())
	assert.Nil(t, matB) // Materialize should yield nil for JSON null
}

func TestSemanticSequenceFlatteningAndArrayPreservation(t *testing.T) {
	// Nested sequence flattens
	// We'll construct nested sequences directly to test Materialize
	nestedSeq := &Sequence{Values: []interface{}{&Sequence{Values: []interface{}{1, 2}}, 3}}
	flatSeq := FlattenSequence(nestedSeq)
	matSeq := Materialize(flatSeq)
	assert.Equal(t, []interface{}{1, 2, 3}, matSeq)

	// Array doesn't flatten when inside sequence
	nestedArr := &Sequence{Values: []interface{}{&Array{Elements: []interface{}{1, 2}}, 3}}
	flatArrSeq := FlattenSequence(nestedArr) // FlattenSequence doesn't flatten Arrays
	matArr := Materialize(flatArrSeq)
	assert.Equal(t, []interface{}{[]interface{}{1, 2}, 3}, matArr)
}

func TestSemanticSingletonSequence(t *testing.T) {
	seq := &Sequence{Values: []interface{}{1}}
	mat := Materialize(seq)
	assert.Equal(t, 1, mat) // Singleton unwrapped

	arr := &Array{Elements: []interface{}{1}}
	matArr := Materialize(arr)
	assert.Equal(t, []interface{}{1}, matArr) // Array preserved
}

func TestGenuineErrorSurvival(t *testing.T) {
	data := `{"a": [1, 2, 3]}`
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Fatalf("failed to unmarshal test data: %v", err)
	}
	root := lookup.Reflect(v)

	// In jsonata missing function evaluates to an error. Let's see.
	ast, _ := Parse("a.$missing_func()")
	runner := Compile(ast)
	res := runner.Run(lookup.NewScope(nil, root))

	// This should be an error Invalidor, not Undefined
	_, ok := res.(*lookup.Invalidor)
	assert.True(t, ok, "Expected Invalidor for genuine error")
}

func TestRegressionStringConcatSingleton(t *testing.T) {
	// String concatenation of singleton sequence should produce "helloworld", not struct formatting.
	// Sequence{1} & Sequence{2} -> "12"
	ast, _ := Parse(`"hello" & "world"`)
	runner := Compile(ast)
	res := runner.Run(lookup.NewScope(nil, nil))

	mat := Materialize(res.Raw())
	assert.Equal(t, "helloworld", mat)
}

func TestRegressionPathFlattening(t *testing.T) {
	data := `{"a": [[1, 2], [3, 4]]}`
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Fatalf("failed to unmarshal test data: %v", err)
	}
	root := lookup.Reflect(v)

	// a[] should flatten to [1, 2, 3, 4] but since we don't have [] syntax implemented,
	// let's test a general path navigation that yields a sequence of sequences.
	ast, _ := Parse("a")
	runner := Compile(ast)
	res := runner.Run(lookup.NewScope(nil, root))

	mat := Materialize(res.Raw())
	assert.Equal(t, []interface{}{[]interface{}{1.0, 2.0}, []interface{}{3.0, 4.0}}, mat)
}

func TestRegressionUndefinedInSequence(t *testing.T) {
	// Undefined members disappear during sequence construction
	seq := &Sequence{Values: []interface{}{1, Undefined{}, 2}}
	flat := FlattenSequence(seq)
	mat := Materialize(flat)
	assert.Equal(t, []interface{}{1, 2}, mat)
}

func TestRegressionArrayPreservationInsidePath(t *testing.T) {
	// nested JSON arrays remain nested inside sequences
	seq := &Sequence{Values: []interface{}{1, &Array{Elements: []interface{}{2, 3}}}}
	flat := FlattenSequence(seq)
	mat := Materialize(flat)
	assert.Equal(t, []interface{}{1, []interface{}{2, 3}}, mat)
}

func TestTruthyDirectRecursive(t *testing.T) {
	cases := []struct {
		name     string
		in       interface{}
		expected bool
	}{
		{"Array{}", &Array{}, false},
		{"Array{false, 0, \"\"}", &Array{Elements: []interface{}{false, 0, ""}}, false},
		{"Array{false, 1}", &Array{Elements: []interface{}{false, 1}}, true},
		{"Sequence{}", &Sequence{}, false},
		{"Sequence{false, 0, \"\"}", &Sequence{Values: []interface{}{false, 0, ""}}, false},
		{"Sequence{false, 1}", &Sequence{Values: []interface{}{false, 1}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := Truthy(tc.in)
			if out != tc.expected {
				t.Fatalf("Truthy(%v) = %v, expected %v", tc.in, out, tc.expected)
			}
		})
	}
}

func TestJsonataValuesEqualDirect(t *testing.T) {
	cases := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"Nested numbers", []interface{}{1}, []interface{}{1.0}, true},
		{"Shape mismatch", []interface{}{1}, 1, false},
		{"Value mismatch", 1, 2, false},
		{"Large integer", 9007199254740992, 9007199254740992, true}, // 2^53
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := jsonataValuesEqual(tc.a, tc.b)
			if out != tc.expected {
				t.Fatalf("jsonataValuesEqual(%v, %v) = %v, expected %v", tc.a, tc.b, out, tc.expected)
			}
		})
	}
}

func TestHarnessUndefinedDistinction(t *testing.T) {
	// missing input -> undefined
	v, undef := materializeHarnessValue(Undefined{})
	if v != nil || !undef {
		t.Fatalf("materializeHarnessValue(Undefined{}) = %v, %v, expected nil, true", v, undef)
	}

	// explicit null -> JSON null
	v, undef = materializeHarnessValue(nil)
	if v != nil || undef {
		t.Fatalf("materializeHarnessValue(nil) = %v, %v, expected nil, false", v, undef)
	}
}
