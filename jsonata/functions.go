package jsonata

import (
	"fmt"

	"github.com/arran4/go-evaluator"
	"github.com/arran4/lookup"
)

var Functions = map[string]evaluator.Function{
	"$substring": &substringFunc{},
}

type substringFunc struct{}

func (s *substringFunc) Call(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("expected at least 2 arguments")
	}

	str, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument 0 must be a string")
	}

	start, ok := lookup.ToInt(args[1])
	if !ok {
		return nil, fmt.Errorf("argument 1 must be an integer")
	}

	length := -1
	if len(args) > 2 {
		l, ok := lookup.ToInt(args[2])
		if ok {
			length = int(l)
		}
	}

	runes := []rune(str)
	if start < 0 {
		start = int64(len(runes)) + start
	}
	if start < 0 {
		start = 0
	}
	if int(start) >= len(runes) {
		return "", nil
	}

	end := int64(len(runes))
	if length != -1 {
		end = start + int64(length)
		if end > int64(len(runes)) {
			end = int64(len(runes))
		}
	}

	if start > end {
		return "", nil
	}

	return string(runes[start:end]), nil
}
