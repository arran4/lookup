package lookup

import (
	"fmt"
	"strings"
)

// ParseSimplePath converts a simple query string like "A.B[0].C" into a Relator
// which can be run against any Pathor. The supported syntax only understands
// dot separated field lookups and integer based indexes using square brackets.
func ParseSimplePath(query string) *Relator {
	r := NewRelator()
	token := strings.Builder{}
	for i := 0; i < len(query); {
		switch query[i] {
		case '.':
			if token.Len() > 0 {
				r = r.Find(token.String())
				token.Reset()
			}
			i++
		case '[':
			if token.Len() > 0 {
				r = r.Find(token.String())
				token.Reset()
			}
			j := strings.IndexByte(query[i:], ']')
			if j == -1 {
				// no closing bracket, treat rest as plain text
				token.WriteString(query[i:])
				i = len(query)
				break
			}
			idx := query[i+1 : i+j]
			r = r.Find("", Index(idx))
			i += j + 1
		default:
			token.WriteByte(query[i])
			i++
		}
	}
	if token.Len() > 0 {
		r = r.Find(token.String())
	}
	return r
}

// QuerySimplePath executes the given simple path query string against the
// provided value using reflection.
func QuerySimplePath(v interface{}, query string) Pathor {
	rel := ParseSimplePath(query)
	root := Reflect(v)
	return rel.Run(NewScope(nil, root))
}

// CompileSimplePath converts a simple query string like "A.B[0].C" into a Relator.
// Unlike ParseSimplePath, it is strict, returning errors for malformed syntax
// and supporting quoted or escaped fields containing metacharacters.
func CompileSimplePath(query string) (*Relator, error) {
	if len(query) == 0 {
		return NewRelator(), nil
	}

	r := NewRelator()
	var inQuote bool
	var inBracket bool
	var escaped bool
	var token strings.Builder

	// We allow an optional leading dot for convenience
	start := 0
	if query[0] == '.' {
		start = 1
	}

	for i := start; i < len(query); i++ {
		c := query[i]

		if escaped {
			token.WriteByte(c)
			escaped = false
			continue
		}

		if inQuote {
			if c == '\\' {
				escaped = true
			} else if c == '"' {
				inQuote = false
				r = r.Find(token.String())
				token.Reset()
				// Expecting a dot or bracket after a quoted key, or EOF
				if i+1 < len(query) && query[i+1] != '.' && query[i+1] != '[' {
					return nil, fmt.Errorf("unexpected character after quoted key at index %d", i+1)
				}
			} else {
				token.WriteByte(c)
			}
			continue
		}

		if inBracket {
			if c == '\\' {
				escaped = true
			} else if c == ']' {
				inBracket = false
				r = r.Find("", Index(token.String()))
				token.Reset()
			} else {
				token.WriteByte(c)
			}
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		switch c {
		case '"':
			if token.Len() > 0 {
				return nil, fmt.Errorf("unexpected quote at index %d", i)
			}
			inQuote = true
		case '[':
			if token.Len() > 0 {
				r = r.Find(token.String())
				token.Reset()
			}
			inBracket = true
		case '.':
			if token.Len() > 0 {
				r = r.Find(token.String())
				token.Reset()
			} else {
				if i > 0 && (query[i-1] == ']' || query[i-1] == '"') {
					// Dot after bracket or quote is allowed
				} else {
					return nil, fmt.Errorf("empty key at index %d", i)
				}
			}
		default:
			token.WriteByte(c)
		}
	}

	if escaped {
		return nil, fmt.Errorf("trailing escape character")
	}
	if inQuote {
		return nil, fmt.Errorf("unmatched quote")
	}
	if inBracket {
		return nil, fmt.Errorf("unmatched bracket")
	}

	if token.Len() > 0 {
		r = r.Find(token.String())
	}

	return r, nil
}
