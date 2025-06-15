package lookup

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

type subFilterFunc struct {
	expression Runner
}

func (s *subFilterFunc) Run(scope *Scope) Pathor {
	b, err := interfaceToBoolOrParse(scope.Position.Raw())
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	if b {
		return scope.Current
	}
	return NewInvalidor(scope.Path(), ErrEvalFail)
}

type filterFunc struct {
	expression Runner
}

func Filter(expression Runner) *filterFunc {
	return &filterFunc{
		expression: expression,
	}
}

func (ef *filterFunc) Run(scope *Scope) Pathor {
	result := arrayOrSliceForEachPath(ExtractPath(scope.Position), nil, scope.Position.Value(), []Runner{
		ef.expression,
		&subFilterFunc{expression: Result()},
	}, scope)
	return result
}

type mapFunc struct {
	expression Runner
}

func Map(expression Runner) *mapFunc {
	return &mapFunc{
		expression: expression,
	}
}

func (ef *mapFunc) Run(scope *Scope) Pathor {
	result := arrayOrSliceForEachPath(ExtractPath(scope.Position), nil, scope.Position.Value(), []Runner{ef.expression}, scope)
	return result
}

func forEach(scope *Scope, v reflect.Value, ef func(pathor Pathor) error) Pathor {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		if v.IsNil() {
			return &Invalidor{
				err: fmt.Errorf("nil element at simple path %s element was %s expected array,slice,map,struct", scope.Path(), "nil"),
			}
		}
		for i := 0; i < v.Len(); i++ {
			f := v.Index(i)
			p := scope.Path() + fmt.Sprintf("[%d]", i)
			if err := ef(&Reflector{
				path: p,
				v:    f,
			}); err != nil {
				return NewInvalidor(p, err)
			}
		}
	}
	return nil
}

type containsFunc struct {
	expression Runner
}

func Contains(runner Runner) *containsFunc {
	return &containsFunc{
		expression: runner,
	}
}

func (ef *containsFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	switch result.Value().Kind() {
	case reflect.Slice, reflect.Array:
		result = arrayOrSliceForEachPath(scope.Path(), nil, scope.Position.Value(), []Runner{
			ValueOf(result),
		}, scope)
		return every(scope, equals(scope, result))
	}
	found := elementOf(result.Value(), scope.Position.Value(), nil)
	return NewConstantor(scope.Path(), found)
}

func In(e Runner) *inFunc {
	return &inFunc{
		expression: e,
	}
}

type inFunc struct {
	expression Runner
}

func (ef *inFunc) Run(scope *Scope) Pathor {
	switch scope.Position.Value().Kind() {
	case reflect.Slice, reflect.Array:
		result := arrayOrSliceForEachPath(scope.Path(), nil, scope.Position.Value(), []Runner{ef}, scope)
		return any(scope, result)
	}
	inThis := ef.expression.Run(scope)
	found := elementOf(scope.Position.Value(), inThis.Value(), nil)
	return NewConstantor(scope.Path(), found)
}

type indexFunc struct {
	i interface{}
}

func Index(i interface{}) *indexFunc {
	return &indexFunc{
		i: i,
	}
}

func (i *indexFunc) Run(scope *Scope) Pathor {
	switch scope.Position.Value().Kind() {
	case reflect.Array, reflect.Slice:
		return evaluateType(scope, scope.Position, i.i)
	default:
		return NewInvalidor(ExtractPath(scope.Position), ErrIndexOfNotArray)
	}
}

func evaluateType(scope *Scope, pathor Pathor, i interface{}) Pathor {
	if i == nil {
		return pathor
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ip, err := interfaceToInt(i)
		if err != nil {
			return NewInvalidor(ExtractPath(pathor), err)
		}
		return arrayOrSlicePath(fmt.Sprintf("%s[%d]", ExtractPath(pathor), ip), ip, pathor.Value())
	case reflect.String:
		simpleValue, err := regexp.Compile(`^-?\d+$`)
		if err != nil {
			return NewInvalidor(ExtractPath(pathor), err)
		}
		if simpleValue.MatchString(i.(string)) {
			ii, err := strconv.ParseInt(i.(string), 10, 64)
			if err != nil {
				return NewInvalidor(fmt.Sprintf("%s[%s]", ExtractPath(pathor), i.(string)), err)
			}
			return arrayOrSlicePath(fmt.Sprintf("%s[%d]", ExtractPath(pathor), ii), ii, pathor.Value())
		}
	case reflect.Struct, reflect.Ptr:
		switch ii := i.(type) {
		case *Constantor:
			return evaluateType(scope, pathor, ii.Raw())
		case Runner:
			p := ii.Run(scope)
			return evaluateType(scope, p, p.Raw())
		default:
			return NewInvalidor(ExtractPath(pathor), ErrIndexValueNotValid)
		}
	default:
		return NewInvalidor(fmt.Sprintf("%s[]", ExtractPath(pathor)), ErrIndexValueNotValid)
	}
	return NewInvalidor(fmt.Sprintf("%s[]", ExtractPath(pathor)), ErrUnknownIndexMode)
}

func Every(e Runner) *everyFunc {
	return &everyFunc{
		expression: e,
	}
}

type everyFunc struct {
	expression Runner
}

func (ef *everyFunc) Run(scope *Scope) Pathor {
	everyThis := ef.expression.Run(scope)
	return every(scope, everyThis)
}

func every(scope *Scope, everyThis Pathor) Pathor {
	v := everyThis.Value()
	every := true
	if err := forEach(scope, v, func(pathor Pathor) error {
		p := Truthy(Result()).Run(scope.Next(pathor))
		b, err := interfaceToBoolOrParse(p.Raw())
		if err == nil && b {
			every = false
		}
		return nil
	}); err != nil {
		return err
	}
	return NewConstantor(scope.Path(), every)
}

func Any(e Runner) *anyFunc {
	return &anyFunc{
		expression: e,
	}
}

type anyFunc struct {
	expression Runner
}

func (ef *anyFunc) Run(scope *Scope) Pathor {
	anyThis := ef.expression.Run(scope)
	return any(scope, anyThis)
}

func any(scope *Scope, anyThis Pathor) Pathor {
	v := anyThis.Value()
	found := false
	if err := forEach(scope, v, func(pathor Pathor) error {
		p := truthy(scope, pathor)
		b, err := interfaceToBoolOrParse(p.Raw())
		if err == nil && b {
			found = true
		}
		return nil
	}); err != nil {
		return err
	}
	return NewConstantor(scope.Path(), found)
}

// TODO Map
// TODO Union
// TODO Intersection
// TODO First - warp index
// TODO Last - wrap index
// TODO Range
