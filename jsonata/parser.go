package jsonata

import (
	"fmt"
	"strconv"
)

// Parse converts a JSONata expression into an AST.
func Parse(expr string) (*AST, error) {
	p := &parser{s: expr}
	node, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if err := p.consumeWhitespace(); err == nil && p.i < len(p.s) {
		return nil, fmt.Errorf("unexpected token at position %d: %c", p.i, p.s[p.i])
	}

	return &AST{Node: node}, nil
}

type parser struct {
	s string
	i int
}

func (p *parser) parseExpression() (Node, error) {
	lhs, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	if err := p.consumeWhitespace(); err != nil {
		return lhs, nil // End of string is fine
	}

	// Check for binary operators
	if p.peek() == '&' {
		p.i++ // consume '&'
		rhs, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &BinaryNode{
			Operator: "&",
			Left:     lhs,
			Right:    rhs,
		}, nil
	}

	if p.peek() == '+' {
		p.i++
		rhs, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &BinaryNode{
			Operator: "+",
			Left:     lhs,
			Right:    rhs,
		}, nil
	}

	return lhs, nil
}

func (p *parser) parseTerm() (Node, error) {
	if err := p.consumeWhitespace(); err != nil {
		return nil, err
	}

	// Parentheses
	if p.peek() == '(' {
		p.i++ // consume '('
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.consumeWhitespace(); err != nil {
			return nil, err
		}
		if p.peek() != ')' {
			return nil, fmt.Errorf("expected )")
		}
		p.i++ // consume ')'

		// Check for following path components (e.g. `(expr).foo` or `(expr)[0]`)
		// We can handle this by treating the parenthesized expr as the "start" of a path?
		// But `parsePath` constructs a `PathNode` with a list of steps.
		// If we return `expr` node, we can't append steps to it easily if it's not a PathNode.

		// If next char is `.` or `[`, we are continuing a path.
		// `parsePath` logic usually starts with Ident.
		// But here we have an arbitrary Node as start.

		// If I encounter `.` or `[` after a term, I should probably wrap the term in a structure that allows further navigation?
		// Or maybe `parsePath` should accept an optional "start node"?
		// No, `PathNode` contains steps. First step can be implicit context?

		// Let's defer this complexity. The failing tests are `foo.(...)`.
		// `foo` is parsed as PathNode.
		// `.` is handled in `parsePath`.
		// `(...)` needs to be handled in `parsePath` as a Step.

		return expr, nil
	}

	// Literal: String
	if p.peek() == '"' || p.peek() == '\'' {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		return &LiteralNode{Value: val}, nil
	}

	// Literal: Number
	if isDigit(p.peek()) || p.peek() == '-' {
		val, err := p.parseValue()
		if err == nil {
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return &LiteralNode{Value: f}, nil
			}
			return &LiteralNode{Value: val}, nil
		}
	}

	// Array Constructor `[...]`
	if p.peek() == '[' {
		p.i++ // consume [
		// Try to parse as list of expressions (literals for now as per previous attempt)
		var litItems []interface{}
		allLiterals := true

		for {
			if err := p.consumeWhitespace(); err != nil {
				return nil, err
			}
			if p.peek() == ']' {
				p.i++
				break
			}

			// Recursive parse
			item, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			if lit, ok := item.(*LiteralNode); ok {
				litItems = append(litItems, lit.Value)
			} else {
				allLiterals = false
				return nil, fmt.Errorf("complex array constructors not supported yet")
			}

			if err := p.consumeWhitespace(); err != nil {
				return nil, err
			}
			if p.peek() == ',' {
				p.i++
			} else if p.peek() != ']' {
				return nil, fmt.Errorf("expected , or ]")
			}
		}

		if allLiterals {
			return &LiteralNode{Value: litItems}, nil
		}
	}

	// Path
	return p.parsePath()
}

func (p *parser) parsePath() (Node, error) {
	var steps []Step

	for {
		if err := p.consumeWhitespace(); err != nil {
			if len(steps) > 0 { break }
			return nil, err
		}

		// Refactored Loop Body:
		var step Step
		var hasStep bool

		// Try Ident
		ident, err := p.parseIdent()
		if err == nil {
			step = Step{Name: ident}
			hasStep = true
		} else {
			// Try SubExpr
			if p.peek() == '(' {
				p.i++
				sub, err := p.parseExpression()
				if err != nil { return nil, err }
				if err := p.consumeWhitespace(); err != nil { return nil, err }
				if p.peek() != ')' { return nil, fmt.Errorf("expected )") }
				p.i++
				step = Step{SubExpr: sub}
				hasStep = true
			}
		}

		if !hasStep {
			if len(steps) > 0 { break }
			return nil, fmt.Errorf("expected identifier or (")
		}

		// Parse Brackets
		if err := p.consumeWhitespace(); err != nil {
			steps = append(steps, step)
			break
		}

		for p.peek() == '[' {
			p.i++
			if err := p.consumeWhitespace(); err != nil { return nil, err }
			if p.peek() == ']' { return nil, fmt.Errorf("empty brackets") }

			if isDigit(p.peek()) || p.peek() == '-' {
				// Index
				start := p.i
				if p.peek() == '-' { p.i++ }
				for isDigit(p.peek()) { p.i++ }
				numStr := p.s[start:p.i]
				if err := p.consumeWhitespace(); err != nil { return nil, err }
				if p.peek() != ']' { return nil, fmt.Errorf("expected ]") }
				p.i++

				num, err := strconv.Atoi(numStr)
				if err != nil { return nil, err }
				step.Index = &num
			} else {
				// Filter
				field, err := p.parseIdent()
				if err != nil { return nil, err }
				if err := p.consumeWhitespace(); err != nil { return nil, err }

				var op string
				switch p.peek() {
				case '=', '>', '<', '!':
					op = string(p.peek())
					p.i++
					if p.peek() == '=' {
						op += string(p.peek())
						p.i++
					}
				default:
					return nil, fmt.Errorf("expected operator")
				}

				if err := p.consumeWhitespace(); err != nil { return nil, err }
				val, err := p.parseValue()
				if err != nil { return nil, err }
				if err := p.consumeWhitespace(); err != nil { return nil, err }
				if p.peek() != ']' { return nil, fmt.Errorf("expected ]") }
				p.i++

				step.Filter = &Predicate{Field: field, Operator: op, Value: val}
			}
			if err := p.consumeWhitespace(); err != nil { break }
		}

		steps = append(steps, step)

		if p.i >= len(p.s) { break }

		if p.peek() == '.' {
			p.i++
			// Continue to next step
		} else {
			// Not a dot, break (could be & or other operator handled by caller)
			break
		}
	}

	return &PathNode{Steps: steps}, nil
}


func (p *parser) consumeWhitespace() error {
	for p.i < len(p.s) {
		c := p.s[p.i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			p.i++
			continue
		}
		if c == '/' && p.i+1 < len(p.s) && p.s[p.i+1] == '*' {
			p.i += 2
			for p.i+1 < len(p.s) && !(p.s[p.i] == '*' && p.s[p.i+1] == '/') {
				p.i++
			}
			if p.i+1 >= len(p.s) {
				return fmt.Errorf("unclosed comment")
			}
			p.i += 2
			continue
		}
		break
	}
	return nil
}

func (p *parser) parseIdent() (string, error) {
	start := p.i
	for p.i < len(p.s) {
		c := p.s[p.i]
		if c == '_' || c == '$' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (p.i > start && c >= '0' && c <= '9') {
			p.i++
			continue
		}
		break
	}
	if start == p.i {
		return "", fmt.Errorf("expected identifier at %d", start)
	}
	return p.s[start:p.i], nil
}

func (p *parser) parseValue() (string, error) {
	if p.peek() == '\'' || p.peek() == '"' {
		quote := p.s[p.i]
		p.i++
		start := p.i
		for p.i < len(p.s) && p.s[p.i] != quote {
			p.i++
		}
		if p.i >= len(p.s) {
			return "", fmt.Errorf("unterminated string")
		}
		val := p.s[start:p.i]
		p.i++ // consume closing quote
		return val, nil
	}
	start := p.i
	for p.i < len(p.s) && (isDigit(p.s[p.i]) || p.s[p.i] == '.') {
		p.i++
	}
	if start == p.i {
		return "", fmt.Errorf("expected value")
	}
	return p.s[start:p.i], nil
}

func (p *parser) peek() byte {
	if p.i >= len(p.s) {
		return 0
	}
	return p.s[p.i]
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
