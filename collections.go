package lookup

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

var simpleIntRegex = regexp.MustCompile(`^-?\d+$`)

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
	v := scope.Position.Value()
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return evaluateType(scope, scope.Position, i.i)
	default:
		// Attempt to treat single item as array of length 1 ?
		// Only if index is 0
		if iIsZero(i.i) {
			// Return the item itself?
			// The path should update to include [0]
			// We can use arrayOrSlicePath logic but applied to single item?
			// Or just return scope.Position (the item)?
			// If we return scope.Position, the path doesn't change?
			// But Index("0") usually implies selecting the 0th element.
			// If Raw() is struct, [0] fails.
			// If we allow it, it should act like it picked the item.
			// But we might need to wrap it in Reflector with updated path?
			p := ExtractPath(scope.Position) + "[0]"
			// Handle if it's already a Pathor
			return &Reflector{
				path: p,
				v:    v,
			}
		}

		return NewInvalidor(ExtractPath(scope.Position), ErrIndexOfNotArray)
	}
}

func iIsZero(i interface{}) bool {
	switch v := i.(type) {
	case int:
		return v == 0
	case string:
		return v == "0"
	case int64:
		return v == 0
	}
	// Add other types if needed or use interfaceToInt
	val, err := interfaceToInt(i)
	return err == nil && val == 0
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
		if simpleIntRegex.MatchString(i.(string)) {
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

type unionFunc struct {
	expression Runner
}

func Union(e Runner) *unionFunc {
	return &unionFunc{expression: e}
}

func (u *unionFunc) Run(scope *Scope) Pathor {
	other := u.expression.Run(scope)
	if _, ok := other.(*Invalidor); ok {
		return other
	}
	leftVals := valueToSlice(scope.Position.Value())
	rightVals := valueToSlice(other.Value())
	result := []interface{}{}
	add := func(v reflect.Value) {
		for _, existing := range result {
			if reflect.DeepEqual(existing, v.Interface()) {
				return
			}
		}
		result = append(result, v.Interface())
	}
	for _, v := range leftVals {
		add(v)
	}
	for _, v := range rightVals {
		add(v)
	}
	return &Reflector{path: scope.Path(), v: reflect.ValueOf(result)}
}

type intersectionFunc struct {
	expression Runner
}

func Intersection(e Runner) *intersectionFunc {
	return &intersectionFunc{expression: e}
}

func (i *intersectionFunc) Run(scope *Scope) Pathor {
	other := i.expression.Run(scope)
	if _, ok := other.(*Invalidor); ok {
		return other
	}
	leftVals := valueToSlice(scope.Position.Value())
	rightVals := valueToSlice(other.Value())
	result := []interface{}{}
	for _, lv := range leftVals {
		for _, rv := range rightVals {
			if reflect.DeepEqual(lv.Interface(), rv.Interface()) {
				// ensure not already added
				exists := false
				for _, existing := range result {
					if reflect.DeepEqual(existing, lv.Interface()) {
						exists = true
						break
					}
				}
				if !exists {
					result = append(result, lv.Interface())
				}
			}
		}
	}
	if len(result) == 0 {
		return &Invalidor{err: ErrNoMatchesForQuery, path: scope.Path()}
	}
	return &Reflector{path: scope.Path(), v: reflect.ValueOf(result)}
}

type appendFunc struct {
	expression Runner
}

func Append(e Runner) *appendFunc {
	return &appendFunc{expression: e}
}

func (u *appendFunc) Run(scope *Scope) Pathor {
	other := u.expression.Run(scope)
	if _, ok := other.(*Invalidor); ok {
		return other
	}
	leftVals := valueToSlice(scope.Position.Value())
	rightVals := valueToSlice(other.Value())
	result := []interface{}{}
	add := func(v reflect.Value) {
		result = append(result, v.Interface())
	}
	for _, v := range leftVals {
		add(v)
	}
	for _, v := range rightVals {
		add(v)
	}
	return &Reflector{path: scope.Path(), v: reflect.ValueOf(result)}
}

type firstFunc struct {
	expression Runner
}

func First(e Runner) *firstFunc { return &firstFunc{expression: e} }

func (f *firstFunc) Run(scope *Scope) Pathor {
	v := scope.Position.Value()
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return NewInvalidor(ExtractPath(scope.Position), ErrIndexOfNotArray)
	}
	if v.IsNil() {
		return NewInvalidor(scope.Path(), fmt.Errorf("nil element"))
	}
	for i := 0; i < v.Len(); i++ {
		p := &Reflector{path: fmt.Sprintf("%s[%d]", scope.Path(), i), v: v.Index(i)}
		r := f.expression.Run(scope.Next(p))
		b, err := interfaceToBoolOrParse(r.Raw())
		if err == nil && b {
			return p
		}
	}
	return NewInvalidor(scope.Path(), ErrNoMatchesForQuery)
}

type lastFunc struct {
	expression Runner
}

func Last(e Runner) *lastFunc { return &lastFunc{expression: e} }

func (l *lastFunc) Run(scope *Scope) Pathor {
	v := scope.Position.Value()
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return NewInvalidor(ExtractPath(scope.Position), ErrIndexOfNotArray)
	}
	if v.IsNil() {
		return NewInvalidor(scope.Path(), fmt.Errorf("nil element"))
	}
	for i := v.Len() - 1; i >= 0; i-- {
		p := &Reflector{path: fmt.Sprintf("%s[%d]", scope.Path(), i), v: v.Index(i)}
		r := l.expression.Run(scope.Next(p))
		b, err := interfaceToBoolOrParse(r.Raw())
		if err == nil && b {
			return p
		}
	}
	return NewInvalidor(scope.Path(), ErrNoMatchesForQuery)
}

type rangeFunc struct {
	start interface{}
	end   interface{}
}

func Range(start, end interface{}) *rangeFunc { return &rangeFunc{start: start, end: end} }

func (rf *rangeFunc) Run(scope *Scope) Pathor {
	v := scope.Position.Value()
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return NewInvalidor(ExtractPath(scope.Position), ErrIndexOfNotArray)
	}
	length := v.Len()
	start, err := evalIndex(scope, rf.start, 0)
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	end, err := evalIndex(scope, rf.end, length)
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	if start < 0 {
		start += length
	}
	if end < 0 {
		end += length
	}
	if start < 0 || start > length || end < 0 || end > length || start > end {
		return NewInvalidor(fmt.Sprintf("%s[%d:%d]", scope.Path(), start, end), ErrIndexOutOfRange)
	}
	slice := v.Slice(start, end)
	return &Reflector{path: fmt.Sprintf("%s[%d:%d]", scope.Path(), start, end), v: slice}
}

func valueToSlice(v reflect.Value) []reflect.Value {
	if !v.IsValid() {
		return nil
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			return nil
		}
		res := make([]reflect.Value, v.Len())
		for i := 0; i < v.Len(); i++ {
			res[i] = v.Index(i)
		}
		return res
	default:
		return []reflect.Value{v}
	}
}

func evalIndex(scope *Scope, val interface{}, def int) (int, error) {
	if val == nil {
		return def, nil
	}
	switch v := val.(type) {
	case *Constantor:
		return evalIndex(scope, v.Raw(), def)
	case Runner:
		r := v.Run(scope)
		if _, ok := r.(*Invalidor); ok {
			return 0, ErrEvalFail
		}
		return evalIndex(scope, r.Raw(), def)
	case Pathor:
		return evalIndex(scope, v.Raw(), def)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(i), nil
	default:
		if i, err := interfaceToInt(v); err == nil {
			return i, nil
		}
	}
	return 0, ErrIndexValueNotValid
}
