package lookup

import (
	"fmt"
	"reflect"
)

type sequenceFunc struct {
	start Runner
	end   Runner
}

func Sequence(start, end Runner) *sequenceFunc {
	return &sequenceFunc{
		start: start,
		end:   end,
	}
}

func (s *sequenceFunc) Run(scope *Scope) Pathor {
	startVal, err := evalIndex(scope, s.start, 0)
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	endVal, err := evalIndex(scope, s.end, 0)
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}

	if startVal > endVal {
		return &Reflector{path: scope.Path(), v: reflect.ValueOf([]interface{}{})}
	}

	size := endVal - startVal + 1
	if size > 100000 {
		return NewInvalidor(scope.Path(), fmt.Errorf("sequence too large: %d > 100000", size))
	}

	result := make([]interface{}, size)
	for i := 0; i < size; i++ {
		result[i] = startVal + i
	}

	return &Reflector{path: scope.Path(), v: reflect.ValueOf(result)}
}
