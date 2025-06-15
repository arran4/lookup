package jsonata

import (
	"github.com/arran4/lookup"
)

// Compile converts the AST into a lookup.Relator which implements Runner.
func Compile(ast *AST) *lookup.Relator {
	r := lookup.NewRelator()
	for _, step := range ast.Steps {
		opts := []lookup.Runner{}
		if step.Index != nil {
			opts = append(opts, lookup.Index(*step.Index))
		}
		if step.Filter != nil {
			opts = append(opts, lookup.Filter(
				lookup.This(step.Filter.Field).Find("", lookup.Equals(lookup.Constant(step.Filter.Value))),
			))
		}
		r = r.Find(step.Name, opts...)
	}
	return r
}
