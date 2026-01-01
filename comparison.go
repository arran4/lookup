package lookup

import (
	"fmt"
	"reflect"
)

type greaterThanFunc struct {
	expression Runner
}

func (ef *greaterThanFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	return greaterThan(scope, result)
}

func greaterThan(scope *Scope, result Pathor) Pathor {
	v1 := scope.Position.Raw()
	v2 := result.Raw()

	f1, err1 := interfaceToFloat(v1)
	f2, err2 := interfaceToFloat(v2)

	if err1 == nil && err2 == nil {
		if f1 > f2 {
			return True(scope.Path())
		}
		return False(scope.Path())
	}

	s1, ok1 := v1.(string)
	s2, ok2 := v2.(string)
	if ok1 && ok2 {
		if s1 > s2 {
			return True(scope.Path())
		}
		return False(scope.Path())
	}

	return False(scope.Path())
}

func GreaterThan(e Runner) *greaterThanFunc {
	return &greaterThanFunc{
		expression: e,
	}
}

type lessThanFunc struct {
	expression Runner
}

func (ef *lessThanFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	return lessThan(scope, result)
}

func lessThan(scope *Scope, result Pathor) Pathor {
	v1 := scope.Position.Raw()
	v2 := result.Raw()

	f1, err1 := interfaceToFloat(v1)
	f2, err2 := interfaceToFloat(v2)

	if err1 == nil && err2 == nil {
		if f1 < f2 {
			return True(scope.Path())
		}
		return False(scope.Path())
	}

	s1, ok1 := v1.(string)
	s2, ok2 := v2.(string)
	if ok1 && ok2 {
		if s1 < s2 {
			return True(scope.Path())
		}
		return False(scope.Path())
	}

	return False(scope.Path())
}

func LessThan(e Runner) *lessThanFunc {
	return &lessThanFunc{
		expression: e,
	}
}

type greaterThanOrEqualFunc struct {
	expression Runner
}

func (ef *greaterThanOrEqualFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	v1 := scope.Position.Raw()
	v2 := result.Raw()

	f1, err1 := interfaceToFloat(v1)
	f2, err2 := interfaceToFloat(v2)

	if err1 == nil && err2 == nil {
		if f1 >= f2 {
			return True(scope.Path())
		}
		return False(scope.Path())
	}

	s1, ok1 := v1.(string)
	s2, ok2 := v2.(string)
	if ok1 && ok2 {
		if s1 >= s2 {
			return True(scope.Path())
		}
		return False(scope.Path())
	}
	return False(scope.Path())
}

func GreaterThanOrEqual(e Runner) *greaterThanOrEqualFunc {
	return &greaterThanOrEqualFunc{expression: e}
}

type lessThanOrEqualFunc struct {
	expression Runner
}

func (ef *lessThanOrEqualFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	v1 := scope.Position.Raw()
	v2 := result.Raw()

	f1, err1 := interfaceToFloat(v1)
	f2, err2 := interfaceToFloat(v2)

	if err1 == nil && err2 == nil {
		if f1 <= f2 {
			return True(scope.Path())
		}
		return False(scope.Path())
	}

	s1, ok1 := v1.(string)
	s2, ok2 := v2.(string)
	if ok1 && ok2 {
		if s1 <= s2 {
			return True(scope.Path())
		}
		return False(scope.Path())
	}
	return False(scope.Path())
}

func LessThanOrEqual(e Runner) *lessThanOrEqualFunc {
	return &lessThanOrEqualFunc{expression: e}
}

type notEqualsFunc struct {
	expression Runner
}

func (ef *notEqualsFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	if !reflect.DeepEqual(result.Raw(), scope.Position.Raw()) {
		return True(scope.Path())
	}
	return False(scope.Path())
}

func NotEquals(e Runner) *notEqualsFunc {
	return &notEqualsFunc{expression: e}
}

func interfaceToFloat(i interface{}) (float64, error) {
	if i == nil {
		return 0, fmt.Errorf("nil")
	}
	switch v := i.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case reflect.Value:
		if v.CanInterface() {
			return interfaceToFloat(v.Interface())
		}
	}
	return 0, fmt.Errorf("not a number")
}
