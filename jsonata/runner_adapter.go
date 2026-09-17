package jsonata

import "github.com/arran4/lookup"

// materializeRunner wraps a generic lookup runner and ensures that any
// JSONata-specific semantics (Sequence/Array) are materialized into standard
// Go scalar/slice representations BEFORE passing into generic lookup runners
// (e.g. string concatenation, arithmetic, comparison).
type materializeRunner struct {
	inner lookup.Runner
}

func (r *materializeRunner) Run(scope *lookup.Scope) lookup.Pathor {
	res := r.inner.Run(scope)

	if inv, ok := res.(*lookup.Invalidor); ok {
		return inv
	}

	if res == nil {
		// Missing evaluation in inner path
		return lookup.Reflect(Undefined{})
	}

	raw := res.Raw()
	if _, ok := raw.(Undefined); ok {
		return lookup.Reflect(Undefined{})
	}

	// Convert internal JSONata value (e.g. Sequence) to standard external value
	mat := Materialize(raw)
	return lookup.Reflect(mat)
}
