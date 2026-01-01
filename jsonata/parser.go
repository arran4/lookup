package jsonata

import (
	"fmt"
	"strconv"
)

// Parse converts a JSONata expression into an AST.
// It supports a very small subset of the language:
//   - dot separated field names
//   - array indexes like [0] or [-1]
//   - equality filters like [field="value"]
//   - basic operators like +, >
func Parse(expr string) (*AST, error) {
	p := &parser{s: expr}
	ast, err := p.parse()
	if err != nil {
		return nil, err
	}
	return ast, nil
}

type parser struct {
	s string
	i int
}

func (p *parser) parse() (*AST, error) {
	ast := &AST{}
	if err := p.consumeWhitespace(); err != nil {
		return nil, err
	}
	for {
		name, err := p.parseIdent()
		if err != nil {
			return nil, err
		}
		if err := p.consumeWhitespace(); err != nil {
			return nil, err
		}
		step := Step{Name: name}
		// zero or more brackets
		for p.peek() == '[' {
			p.i++ // consume '['
			if err := p.consumeWhitespace(); err != nil {
				return nil, err
			}
			if p.peek() == ']' {
				return nil, fmt.Errorf("empty brackets")
			}
			if isDigit(p.peek()) || p.peek() == '-' {
				// index
				start := p.i
				if p.peek() == '-' {
					p.i++
				}
				for isDigit(p.peek()) {
					p.i++
				}
				numStr := p.s[start:p.i]

				if err := p.consumeWhitespace(); err != nil {
					return nil, err
				}
				if p.peek() != ']' {
					return nil, fmt.Errorf("expected closing bracket after index")
				}
				p.i++ // consume closing
				if err := p.consumeWhitespace(); err != nil {
					return nil, err
				}

				num, err := strconv.Atoi(numStr)
				if err != nil {
					return nil, fmt.Errorf("invalid index %s", numStr)
				}
				step.Index = &num
			} else {
				field, err := p.parseIdent()
				if err != nil {
					return nil, err
				}
				if err := p.consumeWhitespace(); err != nil {
					return nil, err
				}

				// Parse operator
				var operator string
				switch p.peek() {
				case '=', '>', '<', '!':
					operator = string(p.peek())
					p.i++
					if p.peek() == '=' { // >=, <=, !=
						operator += string(p.peek())
						p.i++
					}
				default:
					return nil, fmt.Errorf("expected operator after %s", field)
				}

				if err := p.consumeWhitespace(); err != nil {
					return nil, err
				}
				val, err := p.parseValue()
				if err != nil {
					return nil, err
				}
				if err := p.consumeWhitespace(); err != nil {
					return nil, err
				}
				if p.peek() != ']' {
					return nil, fmt.Errorf("expected closing bracket")
				}
				p.i++
				if err := p.consumeWhitespace(); err != nil {
					return nil, err
				}
				step.Filter = &Predicate{Field: field, Operator: operator, Value: val}
			}
		}
		ast.Steps = append(ast.Steps, step)
		if p.i >= len(p.s) {
			break
		}

		// Check for operators (like +) or dot
		if p.s[p.i] == '.' {
			p.i++
			if err := p.consumeWhitespace(); err != nil {
				return nil, err
			}
		} else if p.s[p.i] == '+' {
			p.i++
			if err := p.consumeWhitespace(); err != nil {
				return nil, err
			}
			val, err := p.parseValue()
			if err == nil {
				litStep := Step{Value: val, IsLiteral: true, Operator: "+"}
				ast.Steps = append(ast.Steps, litStep)
				if err := p.consumeWhitespace(); err != nil {
					return nil, err
				}
				if p.i >= len(p.s) {
					break
				}
				if p.s[p.i] == '.' {
					p.i++
					if err := p.consumeWhitespace(); err != nil {
						return nil, err
					}
					continue
				} else {
					// Assume end or another operator (not handled yet)
					break
				}
			} else {
				continue
			}
		} else {
			return nil, fmt.Errorf("expected '.' or operator at position %d, got %c", p.i, p.s[p.i])
		}
	}
	return ast, nil
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
	for p.i < len(p.s) && (isDigit(p.s[p.i]) || p.s[p.i] == '-') {
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
