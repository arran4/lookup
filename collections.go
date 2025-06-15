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
	v := scope.Position.Value()
	found := false
	if err := forEach(scope, v, func(pathor Pathor) error {
		p := Equals(NewConstantor(ExtractPath(result), result.Raw())).Run(scope.Next(pathor))
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

func In(e Runner) *inFunc {
	return &inFunc{
		expression: e,
	}
}

type inFunc struct {
	expression Runner
}

func (ef *inFunc) Run(scope *Scope) Pathor {
	var result Pathor
	switch scope.Position.Value().Kind() {
	case reflect.Slice, reflect.Array:
		result = arrayOrSliceForEachPath(scope.Path(), nil, scope.Position.Value(), []Runner{ef}, scope)
		return any(scope, result)
	default:
		result = scope.Position
	}
	inThis := ef.expression.Run(scope)
	inV := inThis.Value()
	found := false
	if err := forEach(scope, inV, func(pathor Pathor) error {
		p := Equals(NewConstantor(ExtractPath(pathor), result.Raw())).Run(scope.Next(pathor))
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
	return every(scope, everyThis)
}

func every(scope *Scope, everyThis Pathor) Pathor {
	v := everyThis.Value()
	result := scope.Current
	every := true
	if err := forEach(scope, v, func(pathor Pathor) error {
		p := Truthy(NewConstantor(ExtractPath(result), result.Raw())).Run(scope.Next(pathor))
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

type unionFunc struct {
	other Runner
}

func Union(r Runner) *unionFunc {
	return &unionFunc{other: r}
}

func (u *unionFunc) Run(scope *Scope) Pathor {
	left := scope.Position
	right := u.other.Run(scope)
	lv := left.Value()
	rv := right.Value()
	if lv.Kind() != reflect.Array && lv.Kind() != reflect.Slice {
		return NewInvalidor(ExtractPath(left), ErrIndexOfNotArray)
	}
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return NewInvalidor(ExtractPath(right), ErrIndexOfNotArray)
	}
	result := make([]interface{}, 0, lv.Len()+rv.Len())
	for i := 0; i < lv.Len(); i++ {
		result = append(result, lv.Index(i).Interface())
	}
	for i := 0; i < rv.Len(); i++ {
		vi := rv.Index(i).Interface()
		dup := false
		for _, e := range result {
			if reflect.DeepEqual(e, vi) {
				dup = true
				break
			}
		}
		if !dup {
			result = append(result, vi)
		}
	}
	return NewConstantor(scope.Path(), result)
}

type appendFunc struct {
	other Runner
}

func Append(r Runner) *appendFunc {
	return &appendFunc{other: r}
}

func (a *appendFunc) Run(scope *Scope) Pathor {
	left := scope.Position
	right := a.other.Run(scope)
	lv := left.Value()
	rv := right.Value()
	if lv.Kind() != reflect.Array && lv.Kind() != reflect.Slice {
		return NewInvalidor(ExtractPath(left), ErrIndexOfNotArray)
	}
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return NewInvalidor(ExtractPath(right), ErrIndexOfNotArray)
	}
	result := make([]interface{}, 0, lv.Len()+rv.Len())
	for i := 0; i < lv.Len(); i++ {
		result = append(result, lv.Index(i).Interface())
	}
	for i := 0; i < rv.Len(); i++ {
		result = append(result, rv.Index(i).Interface())
	}
	return NewConstantor(scope.Path(), result)
}

type intersectionFunc struct {
	other Runner
}

func Intersection(r Runner) *intersectionFunc {
	return &intersectionFunc{other: r}
}

func (i *intersectionFunc) Run(scope *Scope) Pathor {
	left := scope.Position
	right := i.other.Run(scope)
	lv := left.Value()
	rv := right.Value()
	if lv.Kind() != reflect.Array && lv.Kind() != reflect.Slice {
		return NewInvalidor(ExtractPath(left), ErrIndexOfNotArray)
	}
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return NewInvalidor(ExtractPath(right), ErrIndexOfNotArray)
	}
	result := make([]interface{}, 0)
	for j := 0; j < lv.Len(); j++ {
		lvj := lv.Index(j).Interface()
		for k := 0; k < rv.Len(); k++ {
			if reflect.DeepEqual(lvj, rv.Index(k).Interface()) {
				result = append(result, lvj)
				break
			}
		}
	}
	return NewConstantor(scope.Path(), result)
}

type firstFunc struct {
	expression Runner
}

func First(r Runner) *firstFunc { return &firstFunc{expression: r} }

func (f *firstFunc) Run(scope *Scope) Pathor {
	v := scope.Position.Value()
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return NewInvalidor(ExtractPath(scope.Position), ErrIndexOfNotArray)
	}
	for i := 0; i < v.Len(); i++ {
		p := &Reflector{path: scope.Path() + fmt.Sprintf("[%d]", i), v: v.Index(i)}
		if f.expression == nil {
			return p
		}
		r := f.expression.Run(scope.Next(p))
		b, err := interfaceToBoolOrParse(r.Raw())
		if err == nil && b {
			return p
		}
	}
	return NewInvalidor(scope.Path()+"[0]", ErrIndexOutOfRange)
}

type lastFunc struct {
	expression Runner
}

func Last(r Runner) *lastFunc { return &lastFunc{expression: r} }

func (l *lastFunc) Run(scope *Scope) Pathor {
	v := scope.Position.Value()
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return NewInvalidor(ExtractPath(scope.Position), ErrIndexOfNotArray)
	}
	for i := v.Len() - 1; i >= 0; i-- {
		p := &Reflector{path: scope.Path() + fmt.Sprintf("[%d]", i), v: v.Index(i)}
		if l.expression == nil {
			return p
		}
		r := l.expression.Run(scope.Next(p))
		b, err := interfaceToBoolOrParse(r.Raw())
		if err == nil && b {
			return p
		}
	}
	return NewInvalidor(scope.Path()+"[-1]", ErrIndexOutOfRange)
}

type rangeFunc struct {
	start interface{}
	end   interface{}
}

func Range(start interface{}, end interface{}) *rangeFunc {
	return &rangeFunc{start: start, end: end}
}

func getIndex(scope *Scope, v reflect.Value, i interface{}) (int, error) {
	if i == nil {
		return 0, nil
	}
	switch ii := i.(type) {
	case Runner:
		p := ii.Run(scope)
		return getIndex(scope, v, p.Raw())
	case *Constantor:
		return getIndex(scope, v, ii.Raw())
	case string:
		if m, _ := regexp.MatchString(`^-?\d+$`, ii); m {
			iv, err := strconv.ParseInt(ii, 10, 64)
			return int(iv), err
		}
		return 0, ErrIndexValueNotValid
	default:
		return interfaceToInt(i)
	}
}

func (r *rangeFunc) Run(scope *Scope) Pathor {
	v := scope.Position.Value()
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return NewInvalidor(ExtractPath(scope.Position), ErrIndexOfNotArray)
	}
	l := v.Len()
	start, err := getIndex(scope, v, r.start)
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	end, err := getIndex(scope, v, r.end)
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	if start < 0 {
		start += l
	}
	if end <= 0 {
		end += l
	}
	if start < 0 || end > l || start > end {
		return NewInvalidor(scope.Path(), ErrIndexOutOfRange)
	}
	slice := v.Slice(start, end)
	p := fmt.Sprintf("%s[%d:%d]", ExtractPath(scope.Position), start, end)
	return &Reflector{path: p, v: slice}
}
