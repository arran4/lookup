package lookup_test

import (
	"testing"

	"github.com/arran4/go-evaluator"
	"github.com/arran4/lookup"
)

type mockRunner struct {
	fn func(scope *lookup.Scope) lookup.Pathor
}

func (m mockRunner) Run(scope *lookup.Scope) lookup.Pathor {
	return m.fn(scope)
}

func TestChainPreservesContext(t *testing.T) {
	ctx := &evaluator.Context{
		Variables: map[string]interface{}{},
	}

	s := lookup.NewScopeWithContext(nil, lookup.Simple("init"), ctx)

	r1 := mockRunner{fn: func(scope *lookup.Scope) lookup.Pathor {
		if scope.Context == nil {
			t.Fatal("r1 lost context pointer")
		}

		// Behavioral assertion: can we actually add something to the context
		// and have it preserved in the next step?
		if scope.Context.Variables == nil {
			scope.Context.Variables = map[string]interface{}{}
		}
		scope.Context.Variables["myKey"] = "myBehavior"

		return lookup.Simple("r1_result")
	}}

	r2 := mockRunner{fn: func(scope *lookup.Scope) lookup.Pathor {
		if scope.Context == nil {
			t.Fatal("r2 lost context pointer")
		}

		if scope.Context.Variables == nil {
			t.Fatal("r2 lost context behavior (Variables map is nil)")
		}

		// Check that the context behavior/data added in r1 is still there
		val, ok := scope.Context.Variables["myKey"]
		if !ok {
			t.Fatal("r2 lost context behavior (key not found)")
		}

		if val != "myBehavior" {
			t.Fatalf("r2 context value got %v, want 'myBehavior'", val)
		}

		return lookup.Simple("r2_result")
	}}

	chain := lookup.Chain(r1, r2)
	res := chain.Run(s)
	if res.Raw() != "r2_result" {
		t.Fatalf("expected r2_result, got %v", res.Raw())
	}
}
