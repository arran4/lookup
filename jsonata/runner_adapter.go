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
		return lookup.NewInvalidor("", lookup.ErrNoSuchPath)
	}

	raw := res.Raw()

	mat := Materialize(raw)
	if _, ok := mat.(Undefined); ok {
		return lookup.NewInvalidor("", lookup.ErrNoSuchPath)
	}
	return lookup.Reflect(mat)
}
