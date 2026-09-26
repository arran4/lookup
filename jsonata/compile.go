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
	case *ArrayNode:
		return compileArray(n)
	case *ObjectNode:
		return compileObject(n)
	}
	return lookup.Error(nil) // Should not happen
}

func compileObject(n *ObjectNode) lookup.Runner {
	var props []objectPropertyRunner
	for _, prop := range n.Properties {
		props = append(props, objectPropertyRunner{
			Key:    prop.Key,
			Runner: compileNode(prop.Value),
		})
	}
	return &jsonataObjectRunner{properties: props}
}

func compileArray(n *ArrayNode) lookup.Runner {
	var elements []lookup.Runner
	for _, el := range n.Elements {
		elements = append(elements, compileNode(el))
	}
	return &jsonataArrayRunner{elements: elements}
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
	case "and":
		return &jsonataAndRunner{left: left, right: right}
	case "or":
		return &jsonataOrRunner{left: left, right: right}
	case "..":
		return &jsonataSequenceRunner{inner: lookup.Sequence(&materializeRunner{inner: left}, &materializeRunner{inner: right})}
	case "&", "+", "-", "*", "/", "%", "=", "!=", ">", "<", ">=", "<=", "in":
		return &jsonataBinaryRunner{
			operator: n.Operator,
			left:     left,
			right:    right,
		}
	}
	// Fallback
	return lookup.Error(fmt.Errorf("unsupported binary operator: %s", n.Operator))
}

func compilePath(n *PathNode) lookup.Runner {
	var r lookup.Runner = nil
	for _, step := range n.Steps {
		// Prepare opts (Filters and Indices)
		opts := []lookup.Runner{}
		if step.Index != nil {
			opts = append(opts, &jsonataSingletonRunner{inner: lookup.Index(*step.Index)})
		}
		if step.Filter != nil {
			var field lookup.Runner = lookup.This()
			if step.Filter.Field != "$" {
				field = &jsonataMapRunner{stepRunner: lookup.This(step.Filter.Field), name: step.Filter.Field}
			}
			predicate := &jsonataBinaryRunner{operator: step.Filter.Operator, left: field, right: lookup.Constant(step.Filter.Value)}
			opts = append(opts, &jsonataFilterRunner{predicate: predicate})
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

			if r == nil {
				r = stepRunner
			} else {
				r = &jsonataChain{first: r, second: stepRunner}
			}

		} else if step.SubExpr != nil {
			// SubExpression step: (expr).
			// We evaluate expr in the current context.
			// For each item in current context?
			// `foo.(a & b)`. For each `foo`, evaluate `a & b`.
			// So we wrap in MapRunner.

			subRunner := compileNode(step.SubExpr)
			stepRunner := applyOpts(subRunner)

			mapRunner := &jsonataMapRunner{
				stepRunner: stepRunner,
				name:       "",
			}

			if r == nil {
				r = mapRunner
			} else {
				r = &jsonataChain{first: r, second: mapRunner}
			}

		} else if step.Name == "$" {
			chainStep := &jsonataChain{
				first:  &rootRunner{},
				second: lookup.Find("", opts...),
			}
			mapRunner := &jsonataMapRunner{
				stepRunner: chainStep,
				name:       "$",
			}

			if r == nil {
				r = mapRunner
			} else {
				r = &jsonataChain{first: r, second: mapRunner}
			}

		} else {
			if step.Name == "" {
				if r == nil {
					r = lookup.Find("", opts...)
				} else {
					r = &jsonataChain{first: r, second: lookup.Find("", opts...)}
				}
			} else {
				stepRunner := lookup.This(step.Name).Find("", opts...)

				mapRunner := &jsonataMapRunner{
					stepRunner: stepRunner,
					name:       step.Name,
				}
				if r == nil {
					r = mapRunner
				} else {
					r = &jsonataChain{first: r, second: mapRunner}
				}
			}
		}
	}
	return r
}
