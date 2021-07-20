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
		return scope.Position
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
		&subFilterFunc{ef.expression},
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

func forEach(scope *Scope, v reflect.Value, ef func(f reflect.Value) Pathor) Pathor {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		if v.IsNil() {
			return &Invalidor{
				err: fmt.Errorf("mil element at simple path %s element was %s expected array,slice,map,struct", scope.Path(), "nil"),
			}
		}
		for i := 0; i < v.Len(); i++ {
			f := v.Index(i)
			if err := ef(f); err != nil {
				return err
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
	v := scope.Value()
	found := false
	if err := forEach(scope, v, func(f reflect.Value) Pathor {
		p := Equals(NewConstantor(ExtractPath(result), result.Raw())).Run(scope.Next(Reflect(f.Interface())))
		b, err := interfaceToBoolOrParse(p.Raw())
		if err != nil {
			return NewInvalidor(scope.Path(), err)
		}
		if b {
			found = true
		}
		return nil
	}); err != nil {
		return err
	}
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
	inThis := ef.expression.Run(scope)
	v := inThis.Value()
	result := scope.Current
	found := false
	if err := forEach(scope, v, func(f reflect.Value) Pathor {
		p := Equals(NewConstantor(ExtractPath(result), result.Raw())).Run(scope.Next(Reflect(f.Interface())))
		b, err := interfaceToBoolOrParse(p.Raw())
		if err != nil {
			return NewInvalidor(scope.Path(), err)
		}
		if b {
			found = true
		}
		return nil
	}); err != nil {
		return err
	}
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
	switch scope.Position.Type().Kind() {
	case reflect.Array, reflect.Slice:
	default:
		return NewInvalidor(ExtractPath(scope.Position), ErrIndexOfNotArray)
	}
	return evaluateType(scope, scope.Position, i.i)
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
		simpleValue, err := regexp.Compile("^-?\\d+$")
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
		case Runner:
			p := ii.Run(scope)
			return evaluateType(scope, p, p.Raw())
		case *Constantor:
			return evaluateType(scope, pathor, ii.Raw())
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
	v := everyThis.Value()
	result := scope.Current
	every := true
	if err := forEach(scope, v, func(f reflect.Value) Pathor {
		p := Truthy(NewConstantor(ExtractPath(result), result.Raw())).Run(scope.Next(Reflect(f.Interface())))
		b, err := interfaceToBoolOrParse(p.Raw())
		if err != nil {
			return NewInvalidor(scope.Path(), err)
		}
		if b {
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
	v := anyThis.Value()
	result := scope.Current
	found := false
	if err := forEach(scope, v, func(f reflect.Value) Pathor {
		p := Truthy(NewConstantor(ExtractPath(result), result.Raw())).Run(scope.Next(Reflect(f.Interface())))
		b, err := interfaceToBoolOrParse(p.Raw())
		if err != nil {
			return NewInvalidor(scope.Path(), err)
		}
		if b {
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
