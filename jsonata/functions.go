package jsonata

import (
	"fmt"

	"github.com/arran4/go-evaluator"
	"github.com/arran4/lookup"
)

func GetStandardFunctions() map[string]evaluator.Function {
	return map[string]evaluator.Function{
		"$substring": &substringFunc{},
		"$sum":       &sumFunc{},
		"$count":     &countFunc{},
		"$max":       &maxFunc{},
		"$min":       &minFunc{},
		"$average":   &averageFunc{},
	}
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

type sumFunc struct{}

func (s *sumFunc) Call(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, nil // Or error? JSONata says returns undefined if empty.
	}
	arg := args[0]
	if arg == nil {
		return nil, nil
	}

	val := 0.0

	// Handle array handling... evaluator passes args as is. If it's a slice:
	// If arg is slice, sum elements.
	// If arg is single number, that's the sum.

	switch v := arg.(type) {
	case []interface{}:
		for _, item := range v {
			f, ok := lookup.ToFloat(item)
			if !ok {
				// JSONata ignores non-numbers? Or errors?
				// "It is an error if ... not a number"
				// Let's try convert.
				return nil, fmt.Errorf("item in array is not a number: %T %v", item, item)
			}
			val += f
		}
	case int:
		val = float64(v)
	case float64:
		val = v
	default:
		f, ok := lookup.ToFloat(v)
		if !ok {
			return nil, fmt.Errorf("argument must be an array of numbers or a number")
		}
		val = f
	}

	return val, nil
}

type countFunc struct{}

func (s *countFunc) Call(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0, nil
	}
	arg := args[0]
	if arg == nil {
		return 0, nil // nil is empty sequence?
	}

	switch v := arg.(type) {
	case []interface{}:
		return len(v), nil
	default:
		// Singleton is count 1
		return 1, nil
	}
}

type maxFunc struct{}

func (s *maxFunc) Call(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, nil
	}
	arg := args[0]
	if arg == nil {
		return nil, nil
	}

	var maxVal *float64

	process := func(f float64) {
		if maxVal == nil || f > *maxVal {
			maxVal = &f
		}
	}

	switch v := arg.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		for _, item := range v {
			f, ok := lookup.ToFloat(item)
			if !ok {
				return nil, fmt.Errorf("item in array is not a number")
			}
			process(f)
		}
	default:
		f, ok := lookup.ToFloat(v)
		if !ok {
			return nil, fmt.Errorf("argument must be an array of numbers or a number")
		}
		process(f)
	}

	if maxVal == nil {
		return nil, nil
	}
	return *maxVal, nil
}

type minFunc struct{}

func (s *minFunc) Call(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, nil
	}
	arg := args[0]
	if arg == nil {
		return nil, nil
	}

	var minVal *float64

	process := func(f float64) {
		if minVal == nil || f < *minVal {
			minVal = &f
		}
	}

	switch v := arg.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		for _, item := range v {
			f, ok := lookup.ToFloat(item)
			if !ok {
				return nil, fmt.Errorf("item in array is not a number")
			}
			process(f)
		}
	default:
		f, ok := lookup.ToFloat(v)
		if !ok {
			return nil, fmt.Errorf("argument must be an array of numbers or a number")
		}
		process(f)
	}

	if minVal == nil {
		return nil, nil
	}
	return *minVal, nil
}

type averageFunc struct{}

func (s *averageFunc) Call(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, nil
	}
	arg := args[0]
	if arg == nil {
		return nil, nil
	}

	sum := 0.0
	count := 0

	switch v := arg.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil, nil
		}
		for _, item := range v {
			f, ok := lookup.ToFloat(item)
			if !ok {
				return nil, fmt.Errorf("item in array is not a number")
			}
			sum += f
			count++
		}
	default:
		f, ok := lookup.ToFloat(v)
		if !ok {
			return nil, fmt.Errorf("argument must be an array of numbers or a number")
		}
		sum = f
		count = 1
	}

	if count == 0 {
		return nil, nil
	}
	return sum / float64(count), nil
}
