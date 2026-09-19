package jsonata

import (
	"encoding/json"
	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	astC, err := Parse("c")
	require.NoError(t, err)
	runnerC := Compile(astC)
	resC := runnerC.Run(lookup.NewScope(nil, root))
	matC := Materialize(resC.Raw())

	assert.IsType(t, Undefined{}, matC)

	// Property with null
	astB, err := Parse("b")
	require.NoError(t, err)
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
	ast, err := Parse("a.$missing_func()")
	require.NoError(t, err)
	runner := Compile(ast)
	res := runner.Run(lookup.NewScope(nil, root))

	// This should be an error Invalidor, not Undefined
	_, ok := res.(*lookup.Invalidor)
	assert.True(t, ok, "Expected Invalidor for genuine error")
}

func TestRegressionStringConcatSingleton(t *testing.T) {
	runner := &jsonataBinaryRunner{operator: "&",
		left:  lookup.Constant(&Sequence{Values: []interface{}{"hello"}}),
		right: lookup.Constant(&Sequence{Values: []interface{}{"world"}}),
	}
	res := runner.Run(lookup.NewScope(nil, nil))
	require.Equal(t, "helloworld", res.Raw())
}

func TestRegressionPathFlattening(t *testing.T) {
	for _, tc := range []struct {
		name, expr  string
		input, want interface{}
	}{
		{"mapped arrays", "rows.v", map[string]interface{}{"rows": []interface{}{map[string]interface{}{"v": []interface{}{1, 2}}, map[string]interface{}{"v": []interface{}{3, 4}}}}, []interface{}{1, 2, 3, 4}},
		{"typed singleton", "rows.v", map[string]interface{}{"rows": []map[string]interface{}{{"v": []int{1}}}}, 1},
		{"typed multiple", "rows.v", map[string]interface{}{"rows": []map[string]interface{}{{"v": []int{1, 2}}, {"v": []int{3, 4}}}}, []interface{}{1, 2, 3, 4}},
		{"nested JSON arrays", "a", map[string]interface{}{"a": []interface{}{[]interface{}{1, 2}, []interface{}{3, 4}}}, []interface{}{[]interface{}{1, 2}, []interface{}{3, 4}}},
		{"explicit Array in mapped sequence", "rows.v", map[string]interface{}{"rows": []map[string]interface{}{{"v": &Array{Elements: []interface{}{&Array{Elements: []interface{}{1, 2}}}}}, {"v": 3}}}, []interface{}{[]interface{}{[]interface{}{1, 2}}, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := Parse(tc.expr)
			require.NoError(t, err)
			res := Compile(ast).Run(lookup.NewScope(nil, lookup.Reflect(tc.input)))
			require.NotNil(t, res)
			require.Equal(t, tc.want, res.Raw()) // public result, without another Materialize call
		})
	}
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
		{"Large integer", json.Number("9007199254740993"), int64(9007199254740993), true},
		{"Adjacent large integer", json.Number("9007199254740993"), int64(9007199254740992), false},
		{"Unsigned integer", json.Number("18446744073709551615"), uint64(18446744073709551615), true},
		{"Float32", json.Number("1.5"), float32(1.5), true},
		{"Nested objects", map[string]interface{}{"v": []interface{}{json.Number("2"), json.Number("3.5")}}, map[string]interface{}{"v": []interface{}{int32(2), float64(3.5)}}, true},
		{"Missing object key", map[string]interface{}{"v": nil}, map[string]interface{}{"w": nil}, false},
		{"Array length", []int{1}, []int{1, 2}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, jsonataValuesEqual(tc.b, tc.a), "reverse comparison")
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
