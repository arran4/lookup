package jsonata

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse converts a JSONata expression into an AST.
// It supports a very small subset of the language:
//   - dot separated field names
//   - array indexes like [0] or [-1]
//   - equality filters like [field="value"]
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
	for {
		name, err := p.parseIdent()
		if err != nil {
			return nil, err
		}
		step := Step{Name: name}
		// zero or more brackets
		for p.peek() == '[' {
			p.i++ // consume '['
			if p.peek() == ']' {
				return nil, fmt.Errorf("empty brackets")
			}
			if isDigit(p.peek()) || p.peek() == '-' {
				// index
				numStr := p.readUntil(']')
				p.i++ // consume closing
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
				if p.peek() != '=' {
					return nil, fmt.Errorf("expected '=' after %s", field)
				}
				p.i++
				val, err := p.parseValue()
				if err != nil {
					return nil, err
				}
				if p.peek() != ']' {
					return nil, fmt.Errorf("expected closing bracket")
				}
				p.i++
				step.Filter = &Predicate{Field: field, Value: val}
			}
		}
		ast.Steps = append(ast.Steps, step)
		if p.i >= len(p.s) {
			break
		}
		if p.s[p.i] != '.' {
			return nil, fmt.Errorf("expected '.' at position %d", p.i)
		}
		p.i++
	}
	return ast, nil
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

func (p *parser) readUntil(ch byte) string {
	start := p.i
	for p.i < len(p.s) && p.s[p.i] != ch {
		p.i++
	}
	return strings.TrimSpace(p.s[start:p.i])
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
