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

			field := step.Filter.Field
			var fieldRunner *lookup.Relator
			if field == "$" {
				fieldRunner = lookup.This()
			} else {
				fieldRunner = lookup.This(field)
			}

			opts = append(opts, lookup.Filter(
				fieldRunner.Find("", op),
			))
		}

		if step.Name == "$" {
			r = lookup.Chain(r, &rootRunner{})
			for _, opt := range opts {
				r = lookup.Chain(r, opt)
			}
		} else {
			// Construct the runner for ONE item:
			// Finds name, then applies opts.
			stepRunner := lookup.This(step.Name).Find("", opts...)

			// Wrap in jsonataMapRunner to handle sequence input.
			mapRunner := &jsonataMapRunner{stepRunner: stepRunner}

			// Chain it to previous result
			r = lookup.Chain(r, mapRunner)
		}
	}
	return r
}
