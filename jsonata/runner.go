package jsonata

import (
	"fmt"
	"reflect"

	"errors"
	"github.com/arran4/go-evaluator"
	"github.com/arran4/lookup"
)

type jsonataRunner struct {
	inner lookup.Runner
}

func (r *jsonataRunner) Run(scope *lookup.Scope) lookup.Pathor {
	res := r.inner.Run(scope)

	if inv, ok := res.(*lookup.Invalidor); ok {
		if IsUndefinedError(inv) {
			return lookup.Reflect(Undefined{})
		}
		// In JSONata, querying a non-existent property on a scalar returns undefined instead of erroring
		if isJSONataFieldNoMatch(inv) {
			return lookup.Reflect(Undefined{})
		}
		return inv
	}

	if res == nil {
		return lookup.Reflect(Undefined{})
	}

	raw := res.Raw()
	if _, isUndef := raw.(Undefined); isUndef {
		return lookup.Reflect(Undefined{})
	}

	mat := Materialize(raw)
	return lookup.Reflect(mat)
}

type rootRunner struct{}

func (r *rootRunner) Run(scope *lookup.Scope) lookup.Pathor {
	s := scope
	for s.Parent != nil {
		s = s.Parent
	}
	return s.Current
}

// jsonataMapRunner executes a step on each item of the input if it's a sequence,
// flattening the results. If input is not a sequence, it executes on the input.
type jsonataMapRunner struct {
	stepRunner lookup.Runner
	name       string
}

func (r *jsonataMapRunner) Run(scope *lookup.Scope) lookup.Pathor {
	curr := scope.Current

	if curr == nil || isNilOrNilPointer(curr) {
		return lookup.Reflect(Undefined{})
	}

	raw := curr.Raw()
	if _, ok := raw.(Undefined); ok {
		return lookup.Reflect(Undefined{})
	}

	var items []interface{}

	// Determine iteration strategy
	if seq, ok := raw.(*Sequence); ok {
		items = seq.Values
	} else if arr, ok := raw.(*Array); ok {
		items = arr.Elements
	} else if curr.IsSlice() {
		items, _ = curr.AsSlice()
	} else {
		items = []interface{}{raw}
	}

	var results []interface{}
	for _, item := range items {
		itemPathor := lookup.Reflect(item)
		subScope := scope.Nest(itemPathor)
		res := r.stepRunner.Run(subScope)

		if isNilOrNilPointer(res) {
			continue
		}
		if inv, ok := res.(*lookup.Invalidor); ok {
			if IsUndefinedError(inv) {
				continue
			}
			if isJSONataFieldNoMatch(inv) {
				continue
			}
			if errors.Is(inv, lookup.ErrNoMatchesForQuery) {
				continue
			}
			return inv // real error, stop map evaluation
		}

		resRaw := res.Raw()
		if _, ok := resRaw.(Undefined); ok {
			continue
		}

		resRaw = pathResultValue(resRaw)
		results = append(results, resRaw)
	}

	if len(results) == 0 {
		return lookup.Reflect(Undefined{})
	}

	flat := FlattenSequence(results...)
	return lookup.Reflect(flat)
}

// jsonataChain is a custom chain runner that uses Nest (setting Current) instead of Next (setting Position).
// This ensures that subsequent steps see the result of the previous step as their 'Current' context.
type jsonataChain struct {
	first  lookup.Runner
	second lookup.Runner
}

func (c *jsonataChain) Run(scope *lookup.Scope) lookup.Pathor {
	res := c.first.Run(scope)

	if isNilOrNilPointer(res) {
		return lookup.Reflect(Undefined{})
	}
	if inv, ok := res.(*lookup.Invalidor); ok {
		if IsUndefinedError(inv) {
			return lookup.Reflect(Undefined{})
		}
		// In JSONata, querying a non-existent property on a scalar returns undefined instead of erroring
		if isJSONataFieldNoMatch(inv) {
			return lookup.Reflect(Undefined{})
		}
		return inv
	}
	if _, ok := res.Raw().(Undefined); ok {
		return lookup.Reflect(Undefined{})
	}

	return c.second.Run(scope.Nest(res))
}

// jsonataSingletonRunner wraps a runner (like Index or Filter) and ensures that if the input context
// is not a sequence (array/slice), it is treated as a singleton array.
type jsonataSingletonRunner struct {
	inner lookup.Runner
}

func (r *jsonataSingletonRunner) Run(scope *lookup.Scope) lookup.Pathor {
	curr := scope.Current
	if isNilOrNilPointer(curr) {
		return r.inner.Run(scope)
	}

	if !curr.IsSlice() {
		// Wrap in singleton slice
		singleton := []interface{}{curr.Raw()}
		pathor := lookup.Reflect(singleton)
		// Update scope to point to singleton
		scope = scope.Nest(pathor)
	}

	return r.inner.Run(scope)
}

type jsonataFunctionRunner struct {
	Name string
	Args []lookup.Runner
}

func (r *jsonataFunctionRunner) Run(scope *lookup.Scope) lookup.Pathor {
	var fn evaluator.Function
	if scope.Context != nil && scope.Context.Functions != nil {
		fn = scope.Context.Functions[r.Name]
	}
	if fn == nil {
		return lookup.NewInvalidor("", fmt.Errorf("function %s not implemented", r.Name))
	}

	args := make([]interface{}, len(r.Args))
	for i, arg := range r.Args {
		res := arg.Run(scope)
		if res == nil {
			args[i] = Undefined{}
		} else {
			raw := res.Raw()
			if _, isUndef := raw.(Undefined); isUndef {
				args[i] = raw
			} else {
				args[i] = Materialize(raw)
			}
		}
	}

	res, err := fn.Call(args...)
	if err != nil {
		return lookup.NewInvalidor("", err)
	}
	return lookup.Reflect(res)
}

func isNilOrNilPointer(i interface{}) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return true
	}
	return false
}

func pathResultValue(raw interface{}) interface{} {
	switch raw.(type) {
	case *Sequence, *Array:
		return raw
	}
	rv := reflect.ValueOf(raw)
	if rv.IsValid() && (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) {
		values := make([]interface{}, rv.Len())
		for i := range values {
			values[i] = rv.Index(i).Interface()
		}
		return &Sequence{Values: values}
	}
	return raw
}
