package lookup_test

import (
	"fmt"
	"github.com/arran4/go-evaluator"
	"github.com/arran4/lookup"
	"github.com/google/go-cmp/cmp"
	"testing"
)

type dummyRunner struct {
	result    lookup.Pathor
	err       error
	runCount  int
	lastScope *lookup.Scope
}

func (r *dummyRunner) Run(scope *lookup.Scope) lookup.Pathor {
	r.runCount++
	r.lastScope = scope
	if r.err != nil {
		return lookup.NewInvalidor(scope.Path(), r.err)
	}
	return r.result
}

func TestNestChain(t *testing.T) {
	ctx := &evaluator.Context{
		Variables: map[string]interface{}{"var1": "val1"},
	}

	firstResult := lookup.NewConstantor("first", "first_val")

	t.Run("advances both Current and Position via Nest", func(t *testing.T) {
		first := &dummyRunner{result: firstResult}
		second := &dummyRunner{result: lookup.NewConstantor("second", "second_val")}

		chain := lookup.NestChain(first, second)

		rootScope := lookup.NewScopeWithContext(nil, lookup.Reflect(map[string]interface{}{}), ctx)
		res := chain.Run(rootScope)

		if diff := cmp.Diff(res.Raw(), "second_val"); diff != "" {
			t.Errorf("NestChain result mismatch (-want +got):\n%s", diff)
		}

		if first.runCount != 1 || second.runCount != 1 {
			t.Errorf("Expected both runners to execute once")
		}

		// Verify nested scope properties
		nestedScope := second.lastScope
		if nestedScope.Current != firstResult {
			t.Errorf("Expected nestedScope.Current to be the first runner's result")
		}
		if nestedScope.Position != firstResult {
			t.Errorf("Expected nestedScope.Position to be the first runner's result")
		}
		if nestedScope.Parent != rootScope {
			t.Errorf("Expected nestedScope.Parent to be the root scope")
		}
		if nestedScope.Context != ctx {
			t.Errorf("Expected evaluator.Context to be preserved")
		}
	})

	t.Run("short-circuits on Invalidor", func(t *testing.T) {
		first := &dummyRunner{err: fmt.Errorf("first failed")}
		second := &dummyRunner{result: lookup.NewConstantor("second", "second_val")}

		chain := lookup.NestChain(first, second)

		rootScope := lookup.NewScopeWithContext(nil, lookup.Reflect(map[string]interface{}{}), ctx)
		res := chain.Run(rootScope)

		inv, ok := res.(*lookup.Invalidor)
		if !ok {
			t.Fatalf("Expected Invalidor, got %T", res)
		}
		if inv.Error() != "first failed" {
			t.Errorf("Expected error 'first failed', got %q", inv.Error())
		}

		if first.runCount != 1 {
			t.Errorf("Expected first runner to execute once")
		}
		if second.runCount != 0 {
			t.Errorf("Expected second runner not to execute")
		}
	})
}
