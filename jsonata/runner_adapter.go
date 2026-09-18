package jsonata

import "github.com/arran4/lookup"

// materializeRunner wraps a generic lookup runner and ensures that any
// JSONata-specific semantics (Sequence/Array) are materialized into standard
// Go scalar/slice representations BEFORE passing into generic lookup runners.
type materializeRunner struct {
	inner lookup.Runner
}

func (r *materializeRunner) Run(scope *lookup.Scope) lookup.Pathor {
	res := r.inner.Run(scope)

	if inv, ok := res.(*lookup.Invalidor); ok {
		return inv
	}

	if res == nil {
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

// jsonataBinaryRunner intercepts Undefined values to enforce JSONata operator rules
// without permanently polluting the general lookup package.
type jsonataBinaryRunner struct {
	operator string
	left     lookup.Runner
	right    lookup.Runner
	generic  lookup.Runner
}

func (r *jsonataBinaryRunner) Run(scope *lookup.Scope) lookup.Pathor {
	lRes := r.left.Run(scope)
	if inv, ok := lRes.(*lookup.Invalidor); ok && !IsUndefinedError(inv) {
		return inv
	}
	rRes := r.right.Run(scope)
	if inv, ok := rRes.(*lookup.Invalidor); ok && !IsUndefinedError(inv) {
		return inv
	}

	lRaw := Materialize(lRes.Raw())
	if inv, ok := lRes.(*lookup.Invalidor); ok && IsUndefinedError(inv) {
		lRaw = Undefined{}
	}
	rRaw := Materialize(rRes.Raw())
	if inv, ok := rRes.(*lookup.Invalidor); ok && IsUndefinedError(inv) {
		rRaw = Undefined{}
	}

	_, lUndef := lRaw.(Undefined)
	_, rUndef := rRaw.(Undefined)

	if r.operator == "&" {
		if lUndef {
			lRaw = ""
		}
		if rUndef {
			rRaw = ""
		}
		return lookup.StringConcat(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
	}

	if r.operator == "+" || r.operator == "-" || r.operator == "*" || r.operator == "/" || r.operator == "%" {
		if lUndef || rUndef {
			return lookup.Reflect(Undefined{})
		}
	}

	if r.operator == "=" || r.operator == "!=" || r.operator == ">" || r.operator == "<" || r.operator == ">=" || r.operator == "<=" || r.operator == "in" {
		if lUndef || rUndef {
			return lookup.Reflect(false)
		}
	}

	return r.generic.Run(scope)
}
