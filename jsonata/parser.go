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

	if err := p.consumeWhitespace(); err != nil {
		return nil, err
	}
	if p.i < len(p.s) {
		return nil, fmt.Errorf("unexpected token at position %d: %c", p.i, p.s[p.i])
	}

	return &AST{Node: node}, nil
}

type parser struct {
	s string
	i int
}

// Precedence levels:
// 1. Expression ( & )
// 2. Additive ( + )
// 3. Term ( path, literal, parens )

func (p *parser) parseExpression() (Node, error) {
	// Level 1: &
	lhs, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	if err := p.consumeWhitespace(); err != nil {
		return nil, err
	}

	// Left-associative: loop
	for p.peek() == '&' {
		p.i++ // consume '&'
		rhs, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		lhs = &BinaryNode{
			Operator: "&",
			Left:     lhs,
			Right:    rhs,
		}
		if err := p.consumeWhitespace(); err != nil {
			return nil, err
		}
	}

	return lhs, nil
}

func (p *parser) parseAdditive() (Node, error) {
	// Level 2: + (and - later)
	lhs, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	if err := p.consumeWhitespace(); err != nil {
		return nil, err
	}

	for p.peek() == '+' {
		p.i++ // consume '+'
		rhs, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		lhs = &BinaryNode{
			Operator: "+",
			Left:     lhs,
			Right:    rhs,
		}
		if err := p.consumeWhitespace(); err != nil {
			return nil, err
		}
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

		return &LiteralNode{Value: litItems}, nil
	}

	// Path
	return p.parsePath()
}

func (p *parser) parsePath() (Node, error) {
	var steps []Step

	for {
		if err := p.consumeWhitespace(); err != nil {
			return nil, err
		}

		// Refactored Loop Body:
		var step Step
		var hasStep bool

		// Try Ident
		ident, err := p.parseIdent()
		if err == nil {
			// Check for function call
			if p.peek() == '(' {
				p.i++ // consume '('
				var args []Node
				if p.peek() != ')' {
					for {
						if err := p.consumeWhitespace(); err != nil { return nil, err }
						arg, err := p.parseExpression()
						if err != nil { return nil, err }
						args = append(args, arg)
						if err := p.consumeWhitespace(); err != nil { return nil, err }
						if p.peek() == ')' { break }
						if p.peek() != ',' { return nil, fmt.Errorf("expected , or )") }
						p.i++ // consume ','
					}
				}
				p.i++ // consume ')'
				step = Step{FunctionCall: &FunctionCallNode{Name: ident, Args: args}}
			} else {
				step = Step{Name: ident}
			}
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
			return nil, err
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
			if err := p.consumeWhitespace(); err != nil { return nil, err }
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
