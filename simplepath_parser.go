package lookup

import "strings"

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
