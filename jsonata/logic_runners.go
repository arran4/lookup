package jsonata

import (
	"github.com/arran4/lookup"
)

type jsonataAndRunner struct {
	left  lookup.Runner
	right lookup.Runner
}

func (r *jsonataAndRunner) Run(scope *lookup.Scope) lookup.Pathor {
	lRes := r.left.Run(scope)
	if inv, ok := lRes.(*lookup.Invalidor); ok {
		if !IsUndefinedError(inv) {
			return inv
		}
		return lookup.Reflect(false)
	}

	lRaw := lRes.Raw()
	if !Truthy(lRaw) {
		return lookup.Reflect(false)
	}

	rRes := r.right.Run(scope)
	if inv, ok := rRes.(*lookup.Invalidor); ok {
		if !IsUndefinedError(inv) {
			return inv
		}
		return lookup.Reflect(false)
	}
	rRaw := rRes.Raw()
	if !Truthy(rRaw) {
		return lookup.Reflect(false)
	}

	return lookup.Reflect(true)
}

type jsonataOrRunner struct {
	left  lookup.Runner
	right lookup.Runner
}

func (r *jsonataOrRunner) Run(scope *lookup.Scope) lookup.Pathor {
	lRes := r.left.Run(scope)
	if inv, ok := lRes.(*lookup.Invalidor); ok {
		if !IsUndefinedError(inv) {
			return inv
		}
	} else {
		lRaw := lRes.Raw()
		if Truthy(lRaw) {
			return lookup.Reflect(true)
		}
	}

	rRes := r.right.Run(scope)
	if inv, ok := rRes.(*lookup.Invalidor); ok {
		if !IsUndefinedError(inv) {
			return inv
		}
		return lookup.Reflect(false)
	}
	rRaw := rRes.Raw()
	if Truthy(rRaw) {
		return lookup.Reflect(true)
	}

	return lookup.Reflect(false)
}
