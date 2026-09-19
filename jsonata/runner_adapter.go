package jsonata

import (
	"github.com/arran4/lookup"
)

// materializeRunner wraps a generic lookup runner and ensures that any
// JSONata-specific semantics (Sequence/Array) are materialized into standard
// Go scalar/slice representations BEFORE passing into generic lookup runners.
type materializeRunner struct {
	inner lookup.Runner
}

func (r *materializeRunner) Run(scope *lookup.Scope) lookup.Pathor {
	raw, inv := jsonataResult(r.inner.Run(scope))
	if inv != nil {
		return inv
	}
	return lookup.Reflect(Materialize(raw))
}

// jsonataBinaryRunner intercepts Undefined values to enforce JSONata operator rules
// without permanently polluting the general lookup package.
type jsonataBinaryRunner struct {
	operator string
	left     lookup.Runner
	right    lookup.Runner
}

func (r *jsonataBinaryRunner) Run(scope *lookup.Scope) lookup.Pathor {
	lRaw, inv := jsonataResult(r.left.Run(scope))
	if inv != nil {
		return inv
	}
	rRaw, inv := jsonataResult(r.right.Run(scope))
	if inv != nil {
		return inv
	}
	lRaw, rRaw = Materialize(lRaw), Materialize(rRaw)

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
		switch r.operator {
		case "+":
			return lookup.Add(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case "-":
			return lookup.Subtract(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case "*":
			return lookup.Multiply(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case "/":
			return lookup.Divide(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case "%":
			return lookup.Modulo(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		}
	}

	if r.operator == "=" || r.operator == "!=" || r.operator == ">" || r.operator == "<" || r.operator == ">=" || r.operator == "<=" || r.operator == "in" {
		if lUndef || rUndef {
			return lookup.Reflect(false)
		}
		switch r.operator {
		case "=":
			return lookup.BinaryEquals(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case "!=":
			return lookup.BinaryNotEquals(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case ">":
			return lookup.BinaryGreaterThan(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case "<":
			return lookup.BinaryLessThan(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case ">=":
			return lookup.BinaryGreaterThanOrEqual(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case "<=":
			return lookup.BinaryLessThanOrEqual(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		case "in":
			return lookup.BinaryIn(lookup.Constant(lRaw), lookup.Constant(rRaw)).Run(scope)
		}
	}

	return lookup.NewInvalidor("", lookup.ErrEvalFail)
}

// jsonataSequenceRunner wraps a generic sequence (like range) and ensures its output is a JSONata Sequence.
type jsonataSequenceRunner struct {
	inner lookup.Runner
}

func (r *jsonataSequenceRunner) Run(scope *lookup.Scope) lookup.Pathor {
	raw, inv := jsonataResult(r.inner.Run(scope))
	if inv != nil {
		return inv
	}
	return lookup.Reflect(pathResultValue(raw))
}

// jsonataFilterRunner evaluates predicates itself because generic Filter merges
// predicate errors and ordinary no-match into the same aggregate error.
type jsonataFilterRunner struct{ predicate lookup.Runner }

func (r *jsonataFilterRunner) Run(scope *lookup.Scope) lookup.Pathor {
	raw, inv := jsonataResult(scope.Position)
	if inv != nil {
		return inv
	}
	var items []interface{}
	switch v := pathResultValue(raw).(type) {
	case Undefined:
		return lookup.Reflect(v)
	case *Sequence:
		items = v.Values
	case *Array:
		items = v.Elements
	default:
		items = []interface{}{v}
	}
	var matches []interface{}
	for _, item := range items {
		value, inv := jsonataResult(r.predicate.Run(scope.Nest(lookup.Reflect(item))))
		if inv != nil {
			return inv
		}
		if Truthy(value) {
			matches = append(matches, item)
		}
	}
	return lookup.Reflect(FlattenSequence(matches...))
}

// jsonataResult normalizes absent runner results at JSONata boundaries while
// preserving explicit null and genuine evaluator errors.
func jsonataResult(res lookup.Pathor) (interface{}, *lookup.Invalidor) {
	if isNilOrNilPointer(res) {
		return Undefined{}, nil
	}
	if inv, ok := res.(*lookup.Invalidor); ok {
		if IsUndefinedError(inv) {
			return Undefined{}, nil
		}
		return nil, inv
	}
	return res.Raw(), nil
}
