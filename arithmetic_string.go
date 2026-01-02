package lookup

import (
	"encoding/json"
	"fmt"
)

type stringConcatFunc struct {
	left  Runner
	right Runner
}

func (sf *stringConcatFunc) Run(scope *Scope) Pathor {
	// JSONata spec: "If an operand is not a string, it is converted to a string"
	// "If an operand is null (or missing), it is not converted to a string "null", but treated as an empty string"

	// Wait, JSONata spec says for string(arg):
	// "If arg is not specified (i.e. this function is invoked with no arguments), then the context value is used as the value of arg."
	// "If arg is specified, then..."
	// "If arg is null, then the empty string is returned."

	// For operator &:
	// "The arguments are converted to strings and concatenated."

	leftRes := sf.left.Run(scope)
	rightRes := sf.right.Run(scope)

	s1 := convertToString(leftRes)
	s2 := convertToString(rightRes)

	return NewConstantor(scope.Path(), s1 + s2)
}

func convertToString(p Pathor) string {
	if p == nil {
		return ""
	}
	if _, ok := p.(*Invalidor); ok {
		return ""
	}
	if p.IsNil() {
		return ""
	}

	if s, err := p.AsString(); err == nil {
		return s
	}

	// Complex types or other scalars
	v := p.Raw()
	if v == nil {
		return ""
	}

	// JSON stringify
	// Note: JSONata uses JSON string representation for arrays/objects.
	// But numbers should be simple format.
	// AsString already handles String.

	if p.IsInt() {
		i, _ := p.AsInt()
		return fmt.Sprintf("%d", i)
	}
	if p.IsFloat() {
		f, _ := p.AsFloat()
		// Go's %g default format should be close enough
		return fmt.Sprintf("%v", f)
	}
	if p.IsBool() {
		b, _ := p.AsBool()
		return fmt.Sprintf("%v", b)
	}

	// Slice or Map -> JSON
	b, err := json.Marshal(v)
	if err == nil {
		return string(b)
	}

	// Fallback
	return fmt.Sprintf("%v", v)
}


func StringConcat(left, right Runner) *stringConcatFunc {
	return &stringConcatFunc{
		left:  left,
		right: right,
	}
}
