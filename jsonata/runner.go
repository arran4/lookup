package jsonata

import (
	"fmt"
	"reflect"
	"github.com/arran4/lookup"
)

type jsonataRunner struct {
	inner lookup.Runner
}

func (r *jsonataRunner) Run(scope *lookup.Scope) lookup.Pathor {
	res := r.inner.Run(scope)
	if isNilOrNilPointer(res) {
		return lookup.NewInvalidor("", fmt.Errorf("result is nil"))
	}

	// Singleton unwrapping: JSONata unwraps single-element arrays resulting from path expressions.
	if res.IsSlice() {
		slice, _ := res.AsSlice()
		if len(slice) == 1 {
			return lookup.Reflect(slice[0])
		}
	}
	return res
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

	if isNilOrNilPointer(curr) {
		return lookup.NewInvalidor(r.name, fmt.Errorf("current context is nil"))
	}
	if curr.IsNil() {
		return curr
	}

	if curr.IsSlice() {
		slice, err := curr.AsSlice()
		if err != nil {
			return lookup.NewInvalidor(lookup.ExtractPath(curr), err)
		}

		var results []interface{}
		for _, item := range slice {
			itemPathor := lookup.Reflect(item)

			// We need to construct a scope where 'Current' is the item.
			subScope := scope.Nest(itemPathor)

			res := r.stepRunner.Run(subScope)

			if !isNilOrNilPointer(res) {
				if _, ok := res.(*lookup.Invalidor); ok {
					continue
				}
				if res.IsNil() {
					continue
				}

				if res.IsSlice() {
					s, _ := res.AsSlice()
					results = append(results, s...)
				} else {
					results = append(results, res.Raw())
				}
			}
		}
		if len(results) == 0 {
			return lookup.NewInvalidor(r.name, fmt.Errorf("nothing found"))
		}
		return lookup.Reflect(results)
	}

	// Not a slice.
	res := r.stepRunner.Run(scope)
	if isNilOrNilPointer(res) {
		return lookup.NewInvalidor(r.name, fmt.Errorf("nothing found"))
	}
	if _, ok := res.(*lookup.Invalidor); ok {
		return res
	}
	return res
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
		return lookup.NewInvalidor("", fmt.Errorf("chain broken"))
	}
	if _, ok := res.(*lookup.Invalidor); ok {
		return res
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
	if curr.IsNil() {
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
	switch r.Name {
	case "$substring":
		return r.substring(scope)
	}
	return lookup.NewInvalidor("", fmt.Errorf("function %s not implemented", r.Name))
}

func (r *jsonataFunctionRunner) substring(scope *lookup.Scope) lookup.Pathor {
	if len(r.Args) < 2 {
		return lookup.NewInvalidor("", fmt.Errorf("expected at least 2 arguments"))
	}
	strRes := r.Args[0].Run(scope)
	startRes := r.Args[1].Run(scope)

	if strRes.IsNil() || startRes.IsNil() {
		return lookup.NewInvalidor("", fmt.Errorf("arguments cannot be nil"))
	}

	str, err := strRes.AsString()
	if err != nil {
		return lookup.NewInvalidor("", err)
	}
	start64, err := startRes.AsInt()
	if err != nil {
		// Try float
		f, err2 := startRes.AsFloat()
		if err2 != nil {
			return lookup.NewInvalidor("", err)
		}
		start64 = int64(f)
	}
	start := int(start64)

	length := -1
	if len(r.Args) > 2 {
		lenRes := r.Args[2].Run(scope)
		if !lenRes.IsNil() {
			l, err := lenRes.AsInt()
			if err == nil {
				length = int(l)
			} else {
				f, err := lenRes.AsFloat()
				if err == nil {
					length = int(f)
				}
			}
		}
	}

	// Adjust logic to match JSONata substring
	// JSONata string indices are likely 0-based, but need to check spec.
	// JSONata spec:
	// - string: The string to be substringed.
	// - start: The position within string of the first character to be returned.
	// - length (optional): The number of characters to be returned.
	//
	// Examples:
	// $substring("Hello World", 0, 5) => "Hello"
	// $substring("Hello World", 6) => "World"

	runes := []rune(str)
	if start < 0 {
		start = len(runes) + start
	}
	if start < 0 {
		start = 0
	}
	if start >= len(runes) {
		return lookup.Reflect("")
	}

	end := len(runes)
	if length != -1 {
		end = start + length
		if end > len(runes) {
			end = len(runes)
		}
	}

	if start > end {
		return lookup.Reflect("")
	}

	return lookup.Reflect(string(runes[start:end]))
}

func isNilOrNilPointer(i interface{}) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return true
	}
	return false
}
