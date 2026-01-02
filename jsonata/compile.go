package jsonata

import (
	"github.com/arran4/lookup"
)

// Compile converts the AST into a lookup.Runner.
func Compile(ast *AST) lookup.Runner {
	return &jsonataRunner{inner: compileInternal(ast)}
}

func compileInternal(ast *AST) lookup.Runner {
	var r lookup.Runner = lookup.NewRelator()
	for _, step := range ast.Steps {
		if step.IsLiteral {
			if step.Operator == "+" {
				r = lookup.Add(r, lookup.Constant(step.Value))
			}
			continue
		}

		// Prepare opts (Filters and Indices)
		// We wrap them in jsonataSingletonRunner to ensure they treat scalars as singleton arrays.
		opts := []lookup.Runner{}
		if step.Index != nil {
			opts = append(opts, &jsonataSingletonRunner{inner: lookup.Index(*step.Index)})
		}
		if step.Filter != nil {
			var op lookup.Runner
			switch step.Filter.Operator {
			case "=":
				op = lookup.Equals(lookup.Constant(step.Filter.Value))
			case "!=":
				op = lookup.NotEquals(lookup.Constant(step.Filter.Value))
			case ">":
				op = lookup.GreaterThan(lookup.Constant(step.Filter.Value))
			case "<":
				op = lookup.LessThan(lookup.Constant(step.Filter.Value))
			case ">=":
				op = lookup.GreaterThanOrEqual(lookup.Constant(step.Filter.Value))
			case "<=":
				op = lookup.LessThanOrEqual(lookup.Constant(step.Filter.Value))
			default:
				op = lookup.Equals(lookup.Constant(step.Filter.Value))
			}

			field := step.Filter.Field
			var fieldRunner *lookup.Relator
			if field == "$" {
				fieldRunner = lookup.This()
			} else {
				fieldRunner = lookup.This(field)
			}

			// Filter runner expects list input. Wrap in singleton runner.
			filterRunner := lookup.Filter(fieldRunner.Find("", op))
			opts = append(opts, &jsonataSingletonRunner{inner: filterRunner})
		}

		if step.Name == "$" {
			// $ refers to the query root.

			// We need to apply opts (predicates) to the root value.
			// Root can be scalar or array. Predicates expect array.
			// We use a chain step that Gets Root, then applies Opts (wrapped in Singleton logic).

			// Note: If opts are empty, lookup.Find("", opts...) does nothing (returns Current).

			chainStep := &jsonataChain{
				first: &rootRunner{},
				second: lookup.Find("", opts...),
			}

			r = &jsonataChain{
				first: r,
				second: &jsonataMapRunner{
					stepRunner: chainStep,
					name: "$",
				},
			}

		} else {
			// Construct the runner for ONE item:
			// Finds name, then applies opts.
			// Note: opts are already wrapped in jsonataSingletonRunner.

			// If step.Name is empty (e.g. `(expr)[0]`), lookup.This("") is identity.
			// But wait, if step.Name is empty, we might not want mapRunner?
			// If we have `(expr)[0]`. `expr` returns `[a,b]`.
			// If we run `mapRunner` with `This("")`.
			// `mapRunner` iterates `[a,b]`.
			// Item a. `This("")` -> a. `Index(0)` -> a.
			// Item b. `This("")` -> b. `Index(0)` -> b.
			// Result `[a,b]`.
			// BUT `[0]` on `[a,b]` should be `a`.

			// So if step.Name is empty, we should SKIP mapRunner?
			// But parser might set Name to ""?
			// If Name is provided, we use mapRunner to navigate/flatten.
			// If Name is empty, it's just Predicates on current context.
			// Predicates on a sequence should apply to the sequence!
			// Predicates on a scalar (treated as singleton) apply to singleton.

			// So: If Name is present, we wrap in mapRunner (to navigate and flatten).
			// If Name is empty, we apply opts DIRECTLY to current context (via SingletonWrap).

			if step.Name == "" {
				// Just apply opts.
				// Since we have multiple opts, we can chain them or use Find("", opts...)
				// Opts are already SingletonWrapped.
				// But wait, Find applies opts sequentially.
				// If we have `[0][1]`.
				// `Find("", Index0, Index1)`.
				// Index0 runs. Result passed to Index1.
				// Correct.

				// We attach to `r`.
				r = &jsonataChain{
					first: r,
					second: lookup.Find("", opts...),
				}
			} else {
				stepRunner := lookup.This(step.Name).Find("", opts...)

				// Wrap in jsonataMapRunner to handle sequence input.
				mapRunner := &jsonataMapRunner{
					stepRunner: stepRunner,
					name: step.Name,
				}

				// Chain it to previous result using our Custom Chain (Nest based)
				r = &jsonataChain{first: r, second: mapRunner}
			}
		}
	}
	return r
}
