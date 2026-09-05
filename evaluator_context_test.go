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
	ctx := &evaluator.Context{}
	called := false
	_ = called

	s := lookup.NewScopeWithContext(nil, lookup.Simple("init"), ctx)

	r1 := mockRunner{fn: func(scope *lookup.Scope) lookup.Pathor {
		if scope.Context == nil {
			t.Fatal("r1 lost context")
		}
		called = true
		return lookup.Simple("r1_result")
	}}

	r2 := mockRunner{fn: func(scope *lookup.Scope) lookup.Pathor {
		if scope.Context == nil {
			t.Fatal("r2 lost context")
		}
		if !called {
			t.Fatal("r2 context behavior failed - r1 not called")
		}
		return lookup.Simple("r2_result")
	}}

	chain := lookup.Chain(r1, r2)
	res := chain.Run(s)
	if res.Raw() != "r2_result" {
		t.Fatalf("expected r2_result, got %v", res.Raw())
	}
}
