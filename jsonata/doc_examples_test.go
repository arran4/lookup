package jsonata

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
)

var sampleData = []byte(`{"foo":{"bar":42,"blah":[{"baz":{"fud":"hello"}},{"baz":{"fud":"world"}},{"bazz":"gotcha"}],"blah.baz":"here"},"bar":98}`)

func loadSample(t *testing.T) interface{} {
	var v interface{}
	if err := json.Unmarshal(sampleData, &v); err != nil {
		t.Fatalf("failed to unmarshal sample dataset: %v", err)
	}
	return v
}

// loadPerson reads the example person dataset used by several
// documentation samples.
func loadPerson(t *testing.T) interface{} {
	path := filepath.Join("testdata", "test-suite", "datasets", "dataset1__INPUT.json")
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("failed to read person dataset: %v", err)
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("failed to unmarshal person dataset: %v", err)
	}
	return v
}

func run(t *testing.T, data interface{}, expr string) interface{} {
	ast, err := Parse(expr)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	q := Compile(ast)
	root := lookup.Reflect(data)
	return q.Run(lookup.NewScope(root, root)).Raw()
}

// TestDocExampleSimpleField verifies selecting a single field
// using dot navigation from the documentation examples.
func TestDocExampleSimpleField(t *testing.T) {
	skipIf(t, testFeatureDotNavigation, "dot navigation")
	data := struct{ Foo string }{Foo: "bar"}
	assert.Equal(t, "bar", run(t, data, "Foo"))
}

// TestDocExampleArrayIndex verifies retrieving an element by
// array index from the documentation examples.
func TestDocExampleArrayIndex(t *testing.T) {
	skipIf(t, testFeatureArrayIndexNavigation, "array index")
	data := struct{ Arr []int }{Arr: []int{1, 2, 3}}
	assert.Equal(t, 2, run(t, data, "Arr[1]"))
}

// TestDocExampleEqualityFilter demonstrates filtering an array
// by equality from the documentation examples.
func TestDocExampleEqualityFilter(t *testing.T) {
	skipIf(t, testFeatureEqualityFilter, "equality filter")
	data := loadSample(t)
	out := run(t, data, "foo.blah[bazz='gotcha'].bazz")
	assert.Equal(t, []string{"gotcha"}, out)
}

// TestDocExampleFunctionSum shows calling a function within an
// expression, guarded by a feature flag until implemented.
func TestDocExampleFunctionSum(t *testing.T) {
	skipIf(t, testFeatureFunctionCalls, "function calls")
	data := loadSample(t)
	out := run(t, data, "$sum(foo.blah[baz].baz.fud)")
	assert.Equal(t, nil, out)
}

// TestDocExampleSurname selects a top level field from the
// person dataset used in the simple queries documentation.
func TestDocExampleSurname(t *testing.T) {
	skipIf(t, testFeatureDotNavigation, "dot navigation")
	data := loadPerson(t)
	assert.Equal(t, "Smith", run(t, data, "Surname"))
}

// TestDocExampleAddressCity selects a nested field from the
// person dataset to demonstrate dotted navigation.
func TestDocExampleAddressCity(t *testing.T) {
	skipIf(t, testFeatureDotNavigation, "dot navigation")
	data := loadPerson(t)
	assert.Equal(t, "Winchester", run(t, data, "Address.City"))
}

// TestDocExamplePhoneNegativeIndex returns the last phone entry
// using a negative array index.
func TestDocExamplePhoneNegativeIndex(t *testing.T) {
	skipIf(t, testFeatureArrayIndexNavigation, "array index")
	data := loadPerson(t)
	out := run(t, data, "Phone[-1].type")
	assert.Equal(t, "mobile", out)
}

// TestDocExamplePredicateGreater demonstrates filtering using a comparison predicate from the documentation.
func TestDocExamplePredicateGreater(t *testing.T) {
	skipIf(t, testFeaturePredicate, "predicate comparison")
	data := loadPerson(t)
	out := run(t, data, "Age > 18")
	assert.Equal(t, true, out)
}

// TestDocExampleNumericAdd demonstrates numeric addition.
func TestDocExampleNumericAdd(t *testing.T) {
	skipIf(t, testFeatureNumericOperators, "numeric operators")
	assert.Equal(t, float64(3), run(t, nil, "1+2"))
}

// TestDocExampleBooleanAnd demonstrates boolean AND logic.
func TestDocExampleBooleanAnd(t *testing.T) {
	skipIf(t, testFeatureBooleanOperators, "boolean operators")
	assert.Equal(t, false, run(t, nil, "true and false"))
}

// TestDocExampleStringLength demonstrates calling the $length string function.
func TestDocExampleStringLength(t *testing.T) {
	skipIf(t, testFeatureStringFunctions, "string functions")
	assert.Equal(t, float64(5), run(t, nil, "$length('hello')"))
}

// TestDocExampleNumericFunctionSum demonstrates the $sum numeric function.
func TestDocExampleNumericFunctionSum(t *testing.T) {
	skipIf(t, testFeatureNumericFunctions, "numeric functions")
	data := loadSample(t)
	assert.Equal(t, nil, run(t, data, "$sum(foo.blah[baz].baz.fud)"))
}

// TestDocExampleAggregationCount demonstrates the $count aggregation function.
func TestDocExampleAggregationCount(t *testing.T) {
	skipIf(t, testFeatureAggregationFunctions, "aggregation functions")
	data := loadPerson(t)
	assert.Equal(t, float64(4), run(t, data, "$count(Phone)"))
}

// TestDocExampleArrayAppend demonstrates the $append array function.
func TestDocExampleArrayAppend(t *testing.T) {
	skipIf(t, testFeatureArrayFunctions, "array functions")
	out := run(t, nil, "$append([1,2],3)")
	assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, out)
}

// TestDocExampleObjectMerge demonstrates merging objects.
func TestDocExampleObjectMerge(t *testing.T) {
	skipIf(t, testFeatureObjectFunctions, "object functions")
	out := run(t, nil, "$merge([{\"a\":1},{\"b\":2}])")
	assert.Equal(t, map[string]interface{}{"a": float64(1), "b": float64(2)}, out)
}

// TestDocExampleHigherOrderMap demonstrates the $map higher-order function.
func TestDocExampleHigherOrderMap(t *testing.T) {
	skipIf(t, testFeatureHigherOrderFunctions, "higher order functions")
	data := loadPerson(t)
	out := run(t, data, "$map(Phone,function($v){$v.number})")
	_ = out
}

// TestDocExampleDateTimeNow demonstrates obtaining the current timestamp.
func TestDocExampleDateTimeNow(t *testing.T) {
	skipIf(t, testFeatureDateTimeFunctions, "date-time functions")
	_ = run(t, nil, "$now()")
}

// TestDocExampleRegexMatch demonstrates regular expression matching.
func TestDocExampleRegexMatch(t *testing.T) {
        skipIf(t, testFeatureRegex, "regex functions")
        out := run(t, nil, "$match('abc','/^a/')")
        _ = out
}

// TestDocExampleWildcardPath demonstrates the wildcard path operator.
func TestDocExampleWildcardPath(t *testing.T) {
       skipIf(t, testFeaturePathOperators, "path operators")
       data := loadPerson(t)
       out := run(t, data, "Phone.type")
       assert.Equal(t, []interface{}{"home", "office", "office", "mobile"}, out)
}

// TestDocExampleSubstring demonstrates the $substring string function.
func TestDocExampleSubstring(t *testing.T) {
       skipIf(t, testFeatureStringFunctions, "string functions")
       out := run(t, nil, "$substring('Hello World', 3, 5)")
       assert.Equal(t, "lo Wo", out)
}

// TestDocExampleRound demonstrates the $round numeric function.
func TestDocExampleRound(t *testing.T) {
       skipIf(t, testFeatureNumericFunctions, "numeric functions")
       out := run(t, nil, "$round(123.456, 2)")
       assert.Equal(t, float64(123.46), out)
}

// TestDocExampleBooleanNot demonstrates the $not boolean function.
func TestDocExampleBooleanNot(t *testing.T) {
       skipIf(t, testFeatureBooleanFunctions, "boolean functions")
       out := run(t, nil, "$not(false)")
       assert.Equal(t, true, out)
}

// TestDocExampleRegexReplace demonstrates the $replace string function using a regex.
func TestDocExampleRegexReplace(t *testing.T) {
       skipIf(t, testFeatureRegex, "regex functions")
       out := run(t, nil, "$replace('John Smith and John Jones','John','Mr')")
       assert.Equal(t, "Mr Smith and Mr Jones", out)
}

// TestDocExampleFromMillis demonstrates converting milliseconds to a timestamp.
func TestDocExampleFromMillis(t *testing.T) {
       skipIf(t, testFeatureDateTimeFunctions, "date-time functions")
       out := run(t, nil, "$fromMillis(0)")
       assert.Equal(t, "1970-01-01T00:00:00.000Z", out)
}
