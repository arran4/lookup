package jsonata

import (
	"fmt"
	"github.com/arran4/lookup"
)

type jsonataRunner struct {
	inner lookup.Runner
}

func (r *jsonataRunner) Run(scope *lookup.Scope) lookup.Pathor {
	res := r.inner.Run(scope)
	if res == nil {
		return nil
	}

	// Singleton unwrapping: JSONata unwraps single-element arrays resulting from path expressions.
	if res.IsSlice() {
		slice, _ := res.AsSlice()
		// fmt.Printf("jsonataRunner: Res is slice len %d: %v\n", len(slice), slice)
		if len(slice) == 1 {
			// fmt.Printf("Unwrapping singleton: %v\n", slice[0])
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
	if curr == nil || curr.IsNil() {
		return curr
	}

	// If it's a Slice, we map.
	if curr.IsSlice() {
		slice, err := curr.AsSlice()
		if err != nil {
			return lookup.NewInvalidor(lookup.ExtractPath(curr), err)
		}

		var results []interface{}
		// fmt.Printf("Mapping %s over slice len %d\n", r.name, len(slice))
		for _, item := range slice {
			itemPathor := lookup.Reflect(item)
			subScope := scope.Next(itemPathor)

			// Debug what we are running on
			// fmt.Printf("  Item %d context: %v\n", i, item)

			res := r.stepRunner.Run(subScope)

			if res != nil {
				if _, ok := res.(*lookup.Invalidor); ok {
					continue
				}
				if res.IsNil() {
					continue // Skip explicit nils too?
				}

				// Flatten results
				if res.IsSlice() {
					s, _ := res.AsSlice()
					// fmt.Printf("  Item %d -> slice %v\n", i, s)
					results = append(results, s...)
				} else {
					// fmt.Printf("  Item %d -> val %v\n", i, res.Raw())
					results = append(results, res.Raw())
				}
			} else {
				// fmt.Printf("  Item %d -> nil\n", i)
			}
		}
		if len(results) == 0 {
			// fmt.Printf("  Map result empty\n")
			// Return Invalidor instead of nil/empty slice if nothing matched?
			// But for arrays, filtering results in empty array?
			// No, JSONata returns undefined if nothing found in path step.
			// But if it's a filter/map that results in empty list, is it undefined or empty list?
			// "If the location path selects no values, the result is undefined."
			return lookup.NewInvalidor(r.name, fmt.Errorf("nothing found"))
		}
		return lookup.Reflect(results)
	}

	// Not a slice, run directly
	// fmt.Printf("Running %s on single item %v\n", r.name, curr.Raw())
	res := r.stepRunner.Run(scope)
	if res != nil {
		if _, ok := res.(*lookup.Invalidor); ok {
			return res
		}
	}
	return res
}
