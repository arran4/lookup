package lookup

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

type filterFunc struct {
	expression Runner
}

type subFilterFunc struct {
	expression Runner
}

func (s subFilterFunc) Run(scope *Scope, position Pathor) Pathor {
	b, err := interfaceToBoolOrParse(position.Raw())
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	if b {
		return position
	}
	return NewInvalidor(scope.Path(), ErrEvalFail)
}

func Filter(expression Runner) *filterFunc {
	return &filterFunc{
		expression: expression,
	}
}

func (ef *filterFunc) Run(scope *Scope, position Pathor) Pathor {
	result := arrayOrSliceForEachPath(ExtractPath(position), nil, scope.Value(), []Runner{
		&subFilterFunc{ef.expression},
	}, scope)
	// TODO filter not map
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

func (ef *mapFunc) Run(scope *Scope, position Pathor) Pathor {
	result := arrayOrSliceForEachPath(ExtractPath(position), nil, scope.Value(), []Runner{ef.expression}, scope)
	return result
}

type containsFunc struct {
	expression Runner
}

func Contains(runner Runner) *containsFunc {
	return &containsFunc{
		expression: runner,
	}
}

func (ef *containsFunc) Run(scope *Scope, position Pathor) Pathor {
	result := ef.expression.Run(scope, position)
	v := scope.Value()
	found := false
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		if v.IsNil() {
			return &Invalidor{
				err: fmt.Errorf("mil element at simple path %s element was %s expected array,slice,map,struct", scope.Path(), "nil"),
			}
		}
		for i := 0; i < v.Len(); i++ {
			f := v.Index(i)
			p := Equals(NewConstantor(ExtractPath(result), result.Raw())).Run(scope, Reflect(f.Interface()))
			b, err := interfaceToBoolOrParse(p.Raw())
			if err != nil {
				return NewInvalidor(scope.Path(), err)
			}
			if b {
				found = true
			}
		}
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

func (ef *inFunc) Run(scope *Scope, position Pathor) Pathor {
	inThis := ef.expression.Run(scope, position)
	v := inThis.Value()
	result := scope.Current
	found := false
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		if v.IsNil() {
			return &Invalidor{
				err: fmt.Errorf("mil element at simple path %s element was %s expected array,slice,map,struct", scope.Path(), "nil"),
			}
		}
		for i := 0; i < v.Len(); i++ {
			f := v.Index(i)
			p := Equals(NewConstantor(ExtractPath(result), result.Raw())).Run(scope, Reflect(f.Interface()))
			b, err := interfaceToBoolOrParse(p.Raw())
			if err != nil {
				return NewInvalidor(scope.Path(), err)
			}
			if b {
				found = true
			}
		}
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

func (i *indexFunc) Run(scope *Scope, pathor Pathor) Pathor {
	switch pathor.Type().Kind() {
	case reflect.Array, reflect.Slice:
	default:
		return NewInvalidor(ExtractPath(pathor), ErrIndexOfNotArray)
	}
	return evaluateType(scope, pathor, i.i)
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
			p := ii.Run(scope, pathor)
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

// TODO Map
// TODO Union
// TODO Intersection
// TODO Any
// TODO First - warp index
// TODO Last - wrap index
