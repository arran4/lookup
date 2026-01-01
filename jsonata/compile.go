package jsonata

import (
	"github.com/arran4/lookup"
)

// Compile converts the AST into a lookup.Runner.
func Compile(ast *AST) lookup.Runner {
	var r lookup.Runner = lookup.NewRelator()
	for _, step := range ast.Steps {
		if step.IsLiteral {
			if step.Operator == "+" {
				r = lookup.Add(r, lookup.Constant(step.Value))
			}
			continue
		}

		// Create a runner for the current step
		opts := []lookup.Runner{}
		if step.Index != nil {
			opts = append(opts, lookup.Index(*step.Index))
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
			opts = append(opts, lookup.Filter(
				lookup.This(step.Filter.Field).Find("", op),
			))
		}

		if rel, ok := r.(*lookup.Relator); ok {
			// Optimized path: use Relator.Find which chains internally
			r = rel.Find(step.Name, opts...)
		} else {
			// Chain existing runner with new navigation
			// lookup.This(step.Name) creates a Relator starting from the input
			// We attach opts to it.
			next := lookup.This(step.Name).Find("", opts...)
			r = lookup.Chain(r, next)
		}
	}
	return r
}
