package jsonata

import (
	"fmt"

	"github.com/arran4/lookup"
)

// Compile converts the AST into a lookup.Runner.
func Compile(ast *AST) lookup.Runner {
	return &jsonataRunner{inner: compileNode(ast.Node)}
}

func compileNode(node Node) lookup.Runner {
	switch n := node.(type) {
	case *PathNode:
		return compilePath(n)
	case *BinaryNode:
		return compileBinary(n)
	case *LiteralNode:
		return lookup.Constant(n.Value)
	case *FunctionCallNode:
		return compileFunctionCall(n)
	}
	return lookup.Error(nil) // Should not happen
}

func compileFunctionCall(n *FunctionCallNode) lookup.Runner {
	var args []lookup.Runner
	for _, arg := range n.Args {
		args = append(args, compileNode(arg))
	}

	// Function resolution is done at runtime via Scope/Context
	return &jsonataFunctionRunner{Name: n.Name, Args: args}
}

func compileBinary(n *BinaryNode) lookup.Runner {
	left := compileNode(n.Left)
	right := compileNode(n.Right)

	switch n.Operator {
	case "&":
		return lookup.StringConcat(left, right)
	case "+":
		return lookup.Add(left, right)
	case "-":
		return lookup.Subtract(left, right)
	case "*":
		return lookup.Multiply(left, right)
	case "/":
		return lookup.Divide(left, right)
	case "%":
		return lookup.Modulo(left, right)
	case "=":
		return lookup.BinaryEquals(left, right)
	case "!=":
		return lookup.BinaryNotEquals(left, right)
	case ">":
		return lookup.BinaryGreaterThan(left, right)
	case "<":
		return lookup.BinaryLessThan(left, right)
	case ">=":
		return lookup.BinaryGreaterThanOrEqual(left, right)
	case "<=":
		return lookup.BinaryLessThanOrEqual(left, right)
	case "in":
		return lookup.BinaryIn(left, right)
	case "and":
		return lookup.And(left, right)
	case "or":
		return lookup.Or(left, right)
	case "..":
		return lookup.Sequence(left, right)
	}

	return lookup.Error(fmt.Errorf("unsupported binary operator: %s", n.Operator))
}

func compilePath(n *PathNode) lookup.Runner {
	var r lookup.Runner = lookup.NewRelator()
	for _, step := range n.Steps {
		// Prepare opts (Filters and Indices)
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
			filterRunner := lookup.Filter(fieldRunner.Find("", op))
			opts = append(opts, &jsonataSingletonRunner{inner: filterRunner})
		}

		// Helper to apply opts
		applyOpts := func(base lookup.Runner) lookup.Runner {
			if len(opts) == 0 {
				return base
			}
			// Use Find("", opts...) which chains runners on the result of base.
			// But base needs to be linked.
			// If base is This("name"), Find("", opts) works on that result.
			// But here we return a Runner.

			// We can chain base + Find("", opts...).
			return &jsonataChain{
				first:  base,
				second: lookup.Find("", opts...),
			}
		}

		if step.FunctionCall != nil {
			// Function Call Step
			funcRunner := compileFunctionCall(step.FunctionCall)
			stepRunner := applyOpts(funcRunner)

			// If the function call is part of a path (e.g. foo.bar()),
			// the function is executed.
			// NOTE: In JSONata, functions like $substring are usually global or defined in scope,
			// not methods on objects (unless purely method call syntax which JSONata is loose about).
			// If the step name is empty, it means just apply logic.

			// We should probably treat it similar to SubExpr or simple name but with execution.
			// But wait, step.FunctionCall.Args are expressions evaluated in current scope.

			// If this is part of a path chain, the previous result is current scope.

			r = &jsonataChain{first: r, second: stepRunner}

		} else if step.SubExpr != nil {
			// SubExpression step: (expr).
			// We evaluate expr in the current context.
			// For each item in current context?
			// `foo.(a & b)`. For each `foo`, evaluate `a & b`.
			// So we wrap in MapRunner.

			subRunner := compileNode(step.SubExpr)
			// Apply opts to the result of subExpr? `(expr)[0]`. Yes.

			// Combine subRunner + opts
			stepRunner := applyOpts(subRunner)

			// Wrap in MapRunner to ensure iteration over current context
			// We don't have a "Name" for this step, it's just a mapping.
			// But MapRunner logic relies on nesting.
			mapRunner := &jsonataMapRunner{
				stepRunner: stepRunner,
				name:       "", // Anonymous step
			}

			r = &jsonataChain{first: r, second: mapRunner}

		} else if step.Name == "$" {
			// $ refers to the query root.
			chainStep := &jsonataChain{
				first:  &rootRunner{},
				second: lookup.Find("", opts...),
			}

			r = &jsonataChain{
				first: r,
				second: &jsonataMapRunner{
					stepRunner: chainStep,
					name:       "$",
				},
			}

		} else {
			if step.Name == "" {
				// Just apply opts to current context.
				r = &jsonataChain{
					first:  r,
					second: lookup.Find("", opts...),
				}
			} else {
				stepRunner := lookup.This(step.Name).Find("", opts...)

				mapRunner := &jsonataMapRunner{
					stepRunner: stepRunner,
					name:       step.Name,
				}
				r = &jsonataChain{first: r, second: mapRunner}
			}
		}
	}
	return r
}
