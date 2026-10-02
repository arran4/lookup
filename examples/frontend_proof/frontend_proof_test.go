package frontend_proof

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/arran4/lookup"
)

// A tiny DSL/AST for the proof
type Query struct {
	Steps []Step
}

type Step struct {
	Field string // simple field navigation
	Index *int   // numeric indexing
	Func  string // generic function
}

// compile translates the Query to a lookup.Runner
func (q *Query) Compile() lookup.Runner {
	var current lookup.Runner

	for _, step := range q.Steps {
		var stepRunner lookup.Runner

		if step.Field != "" {
			// Find field
			stepRunner = lookup.Find(step.Field)
		} else if step.Index != nil {
			// Index
			stepRunner = lookup.Find("", lookup.Index(fmt.Sprintf("%d", *step.Index)))
		} else if step.Func != "" {
			// Simple function step
			stepRunner = &ProofFunctionRunner{FuncName: step.Func}
		}

		if current == nil {
			current = stepRunner
		} else {
			// Use lookup.NestChain so that relative lookups in the next stage
			// search from the result of the previous stage.
			current = lookup.NestChain(current, stepRunner)
		}
	}

	return current
}

// ProofFunctionRunner applies a simple transformation
type ProofFunctionRunner struct {
	FuncName string
}

func (p *ProofFunctionRunner) Run(scope *lookup.Scope) lookup.Pathor {
	if p.FuncName == "length" {
		val := scope.Position.Raw()

		switch v := val.(type) {
		case string:
			return lookup.NewConstantor(scope.Path(), len(v))
		case []interface{}:
			return lookup.NewConstantor(scope.Path(), len(v))
		}

		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			return lookup.NewConstantor(scope.Path(), rv.Len())
		}
	}

	return lookup.NewInvalidor(scope.Path(), fmt.Errorf("unknown function %s", p.FuncName))
}

func TestFrontendProof(t *testing.T) {
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice"},
			map[string]interface{}{"name": "Bob"},
			map[string]interface{}{"name": "Charlie"},
		},
	}

	idx := 1
	q := Query{
		Steps: []Step{
			{Field: "users"},
			{Index: &idx}, // should get Bob
			{Field: "name"},
			{Func: "length"}, // length of "Bob" is 3
		},
	}

	runner := q.Compile()
	scope := lookup.NewScope(nil, lookup.Reflect(data))
	res := runner.Run(scope)

	if inv, ok := res.(*lookup.Invalidor); ok {
		t.Fatalf("Query failed: %v", inv.Error())
	}

	val := res.Raw()
	if val != 3 {
		t.Fatalf("Expected 3, got %v", val)
	}
}
