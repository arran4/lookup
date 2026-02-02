package lookup

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// arrayOrSliceForEachPath if the array/slice path isn't found or the path isn't a valid index (ie a string) then this
// function extracts all matches from the array and puts them into a type matched array if possible otherwise a generic
// []interface{} map.
func arrayOrSliceForEachPath(prefix string, paths []string, v reflect.Value, runners []Runner, scope *Scope) Pathor {
	typeCount := map[reflect.Type]int{}
	type Pair struct {
		Boxed   Pathor
		Unboxed Pathor
	}
	result := make([]*Pair, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		p := prefix + fmt.Sprintf("[%d]", i)
		vi := v.Index(i)
		var boxed Pathor
		// Check if the element itself implements Finder/Pathor
		if vi.CanInterface() {
			if f, ok := vi.Interface().(Finder); ok {
				// It's already a Finder/Pathor (most likely Interfaceor or Reflector)
				// If it is a Finder, we can just call Find on it.
				// But we need a Pathor for 'Boxed'. Pathor includes Finder.
				if parthor, ok := f.(Pathor); ok {
					boxed = parthor
				}
			}
		}

		if boxed == nil {
			boxed = &Reflector{
				path: p,
				v:    vi,
			}
		}

		vipath := &Pair{
			Boxed: boxed,
		}
		for _, path := range paths {
			vipath.Boxed = vipath.Boxed.Find(path)
		}
		if _, ok := vipath.Boxed.(*Invalidor); ok {
			continue
		}
		skip := false
		var myScope *Scope
		if scope != nil {
			myScope = scope.Nest(vipath.Boxed)
		} else {
			myScope = NewScope(nil, vipath.Boxed)
		}

		for _, e := range runners {
			ee := e.Run(myScope.Next(vipath.Boxed))
			if _, ok := ee.(*Invalidor); ok {
				skip = true
				continue
			}
			if ee == nil {
				skip = true
				continue
			}
			vipath.Boxed = ee
		}
		if skip {
			continue
		}
		t := vipath.Boxed.Type()
		for e := vipath.Boxed.Value(); e.IsValid(); e = e.Elem() {
			if e.IsValid() && e.Kind() != reflect.Interface {
				t = e.Type()
				vipath.Unboxed = &Reflector{
					path: p,
					v:    e,
				}
				break
			}
		}
		result = append(result, vipath)
		typeCount[t] += 1
	}
	boxing := true
	at := v.Type()
	p := prefix + "[*]"
	for _, path := range paths {
		if len(path) > 0 {
			p = p + "." + path
		}
	}
	switch len(typeCount) {
	case 0:
		err := ErrNoMatchesForQuery
		if len(paths) == 0 && len(runners) > 0 {
			err = ErrEvalFail
		}
		return &Invalidor{
			err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", "int", strings.Join(paths, "."), err),
			path: p,
		}
	case 1:
		for k := range typeCount {
			at = k
			boxing = false
		}
	default:
		for at != nil {
			c := 0
			for t := range typeCount {
				if !t.AssignableTo(at) {
					c++
				}
			}
			if c == 0 {
				break
			}
			switch at.Kind() {
			case reflect.Slice:
				at = at.Elem()
				continue
			case reflect.Array:
				at = at.Elem()
				continue
			case reflect.Map:
				at = at.Elem()
				continue
			case reflect.Ptr:
				at = at.Elem()
				continue
			case reflect.Func:
				at = at.Out(1)
				continue
			}
			var ni interface{} = nil
			at = reflect.TypeOf((*interface{})(&ni)).Elem()
			break
		}
	}
	resultV := reflect.MakeSlice(reflect.SliceOf(at), len(result), len(result))
	for i := 0; i < len(result); i++ {
		if boxing {
			resultV.Index(i).Set(result[i].Boxed.Value())
		} else {
			resultV.Index(i).Set(result[i].Unboxed.Value())
		}
	}
	return &Reflector{
		path: p,
		v:    resultV,
	}
}

// arrayOrSlicePath attempts to extract the path as an index from the array, if this fails it will then use the index to
// assemble an array of matching children using arrayOrSliceForEachPath
func arrayOrSlicePath(prefix string, path interface{}, v reflect.Value) Pathor {
	var i int64
	pathS := "0"
	switch path := path.(type) {
	case string:
		var err error
		pathS = path
		i, err = strconv.ParseInt(path, 10, 64)
		if err != nil {
			p := prefix + "[" + strconv.Quote(pathS) + "]"
			return &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because it couldn't become a integer %w and it wasn't a valid key", "int", path, err),
				path: p,
			}
		}
	default:
		pathI, err := interfaceToInt(path)
		if err != nil {
			p := prefix + "[" + strconv.Quote(pathS) + "]"
			return &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because it couldn't become a integer %w and it wasn't a valid key", "int", path, err),
				path: p,
			}
		}
		i = int64(pathI)
		pathS = fmt.Sprintf("%d", i)
	}
	p := prefix + "[" + strconv.Quote(pathS) + "]"
	l := int64(v.Len())
	if i < 0 {
		i += l
	}
	if i < 0 || i >= l {
		return &Invalidor{
			err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %s", "int", pathS, "it was out of range"),
			path: p,
		}
	}
	rv := v.Index(int(i))
	return &Reflector{
		path: p,
		v:    rv,
	}
}

// mapPath attempts to convert the path to the appropriate from of key if it can be determined then look up the value
// and return it.
func mapPath(prefix string, path string, v reflect.Value) Pathor {
	p := prefix + "." + strconv.Quote(path)
	if prefix == "" || strings.HasSuffix(prefix, ".") {
		p = prefix + strconv.Quote(path)
	}
	//p := prefix + "[\"" + strconv.Quote(path) + "\"]"
	k, pather := extractKey(path, v, p)
	if pather != nil {
		return pather
	}
	ve := v.MapIndex(k)
	for ve.Kind() == reflect.Interface {
		if ve.IsNil() {
			break
		}
		ve = ve.Elem()
	}
	if !ve.IsValid() {
		return &Invalidor{
			err:  fmt.Errorf("element not found at simple path %s element was %s expected %s", p, v.Kind(), v.Type().Key().Kind()),
			path: p,
		}
	}
	return &Reflector{
		path: p,
		v:    ve,
	}
}

// extractKey tries to convert the path into the key type required and return it, or return an error in a Pathor
func extractKey(path string, v reflect.Value, p string) (reflect.Value, Pathor) {
	k := reflect.ValueOf(path)
	kt := v.Type().Key().Kind()
	switch kt {
	case reflect.Bool:
		if v, err := strconv.ParseBool(path); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(v)
		}
	case reflect.Int:
		if v, err := strconv.ParseInt(path, 10, 64); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(int(v))
		}
	case reflect.Int8:
		if v, err := strconv.ParseInt(path, 10, 8); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(int8(v))
		}
	case reflect.Int16:
		if v, err := strconv.ParseInt(path, 10, 16); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(int16(v))
		}
	case reflect.Int32:
		if v, err := strconv.ParseInt(path, 10, 32); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(int32(v))
		}
	case reflect.Int64:
		if v, err := strconv.ParseInt(path, 10, 64); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(int64(v))
		}
	case reflect.Uint:
		if v, err := strconv.ParseInt(path, 10, 64); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(uint(v))
		}
	case reflect.Uint8:
		if v, err := strconv.ParseInt(path, 10, 8); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(uint8(v))
		}
	case reflect.Uint16:
		if v, err := strconv.ParseInt(path, 10, 16); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(uint16(v))
		}
	case reflect.Uint32:
		if v, err := strconv.ParseInt(path, 10, 32); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(uint32(v))
		}
	case reflect.Uint64:
		if v, err := strconv.ParseInt(path, 10, 64); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(uint64(v))
		}
	//case reflect.Uintptr:
	case reflect.Float32:
		if v, err := strconv.ParseFloat(path, 32); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(float32(v))
		}
	case reflect.Float64:
		if v, err := strconv.ParseFloat(path, 64); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(float64(v))
		}
	case reflect.Complex64:
		if v, err := strconv.ParseComplex(path, 64); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(complex64(v))
		}
	case reflect.Complex128:
		if v, err := strconv.ParseComplex(path, 128); err != nil {
			return reflect.Value{}, &Invalidor{
				err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %w", kt, path, err),
				path: p,
			}
		} else {
			k = reflect.ValueOf(complex128(v))
		}
	//case reflect.Array:
	case reflect.Interface:
	//  Interface.. We will just do string
	//case reflect.Map:
	//case reflect.Ptr:
	//case reflect.Slice:
	case reflect.String:
	//case reflect.Struct:
	//case reflect.UnsafePointer:
	default:
		return reflect.Value{}, &Invalidor{
			err:  fmt.Errorf("invalid element at simple path %s element was %s expected %s", p, v.Kind(), v.Type().Key().Kind()),
			path: p,
		}

	}
	return k, nil
}

// structPath attempts to extract a field matching the name provided, if it can't do that then it attempts to look for a
// function and run it if it matches the provided parameters.
func structPath(prefix string, path string, v reflect.Value, pv *reflect.Value) Pathor {
	p := prefix + "." + path
	if unicode.IsLower([]rune(path)[0]) {
		return &Invalidor{
			err:  fmt.Errorf("invalid element at simple path %s element was not found - not exported", p),
			path: p,
		}
	}
	if prefix == "" || strings.HasSuffix(prefix, ".") {
		p = prefix + path
	}
	f := v.FieldByName(path)
	if f.IsValid() {
		return &Reflector{
			path: p,
			v:    f,
		}
	}
	var m reflect.Value
	if pv != nil {
		m = pv.MethodByName(path)
	} else {
		m = v.MethodByName(path)
	}
	pather := runMethod(m, p)
	if pather != nil {
		return pather
	}
	return &Invalidor{
		err:  fmt.Errorf("invalid element at simple path %s field or method was not found", p),
		path: p,
	}
}

// runMethod runs the method that's provided, if the definition is valid and then returns the appropriate Pathor or nil.
func runMethod(m reflect.Value, p string) Pathor {
	if !m.IsValid() {
		return nil
	}
	mt := m.Type()
	outValuePass := true
	switch mt.NumOut() {
	case 2:
		var e error = nil
		errType := reflect.TypeOf((*error)(&e)).Elem()
		mt.Out(1).AssignableTo(errType)
		fallthrough
	case 1:
		switch mt.Out(0).Kind() {
		case reflect.Invalid:
		case reflect.Chan:
		case reflect.Uintptr:
		case reflect.UnsafePointer:
		default:
			outValuePass = true
		}
	}
	if m.IsValid() && mt.NumIn() == 0 && outValuePass {
		p += "()"
		mra := m.Call([]reflect.Value{})
		if len(mra) == 2 && !mra[1].IsNil() {
			err := fmt.Errorf("unknown error")
			if e, ok := mra[1].Interface().(error); ok {
				err = e
			}
			return &Invalidor{
				err:  fmt.Errorf("invalid element at simple path %s method call returned error %w", p, err),
				path: p,
			}
		}
		if len(mra) >= 1 {
			return &Reflector{
				path: p,
				v:    mra[0],
			}
		}
	}
	return nil
}

func elementOf(v reflect.Value, in reflect.Value, pv *reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	if !in.IsValid() {
		return false
	}
	switch in.Kind() {
	case reflect.Array:
		for i := 0; i < in.Len(); i++ {
			f := in.Index(i)
			if reflect.DeepEqual(v.Interface(), f.Interface()) {
				return true
			}
		}
	case reflect.Func:
		r := runMethod(in, "")
		return elementOf(r.Value(), in, nil)
	case reflect.Map:
		for _, k := range in.MapKeys() {
			f := in.MapIndex(k)
			if reflect.DeepEqual(v.Interface(), f.Interface()) {
				return true
			}
		}
	case reflect.Ptr:
		return elementOf(v.Elem(), in.Elem(), &v)
	case reflect.Slice:
		switch in.Type().Elem().Kind() {
		case reflect.Int:
			if val, ok := in.Interface().([]int); ok {
				if v.Kind() == reflect.Int && v.Type() == in.Type().Elem() {
					vv := int(v.Int())
					for _, x := range val {
						if x == vv {
							return true
						}
					}
					return false
				}
			}
		case reflect.Int64:
			if val, ok := in.Interface().([]int64); ok {
				if v.Kind() == reflect.Int64 && v.Type() == in.Type().Elem() {
					vv := v.Int()
					for _, x := range val {
						if x == vv {
							return true
						}
					}
					return false
				}
			}
		case reflect.String:
			if val, ok := in.Interface().([]string); ok {
				if v.Kind() == reflect.String && v.Type() == in.Type().Elem() {
					vv := v.String()
					for _, x := range val {
						if x == vv {
							return true
						}
					}
					return false
				}
			}
		case reflect.Interface:
			if val, ok := in.Interface().([]interface{}); ok {
				vv := v.Interface()
				if v.Type().Comparable() {
					kind := v.Kind()
					if kind != reflect.Float32 && kind != reflect.Float64 && kind != reflect.Complex64 && kind != reflect.Complex128 {
						for _, x := range val {
							if vv == x {
								return true
							}
						}
						return false
					}
				}
				for _, x := range val {
					if reflect.DeepEqual(vv, x) {
						return true
					}
				}
				return false
			}
		}
		for i := 0; i < in.Len(); i++ {
			f := in.Index(i)
			if reflect.DeepEqual(v.Interface(), f.Interface()) {
				return true
			}
		}
	case reflect.Struct:
		for i := 0; i < in.NumField(); i++ {
			f := in.Field(i)
			if reflect.DeepEqual(v.Interface(), f.Interface()) {
				return true
			}
		}
		for i := 0; i < in.NumMethod(); i++ {
			var f reflect.Value
			if pv == nil {
				f = v.Method(i)
			} else {
				f = pv.Method(i)
			}
			fr := runMethod(f, "")
			if elementOf(fr.Value(), in, nil) {
				return true
			}
		}
	default:
		return reflect.DeepEqual(v.Interface(), in.Interface())
	}
	return false
}

func interfaceToInt(i interface{}) (int, error) {
	switch i := i.(type) {
	case int:
		return i, nil
	case int8:
		return int(i), nil
	case int16:
		return int(i), nil
	case int32:
		return int(i), nil
	case int64:
		return int(i), nil
	case uint:
		return int(i), nil
	case uint8:
		return int(i), nil
	case uint16:
		return int(i), nil
	case uint32:
		return int(i), nil
	case uint64:
		return int(i), nil
	case uintptr:
		return int(i), nil
	case float32:
		return int(i), nil
	case float64:
		return int(i), nil
	}
	return 0, errors.New("unknown number type")
}

func interfaceToString(i interface{}) (string, error) {
	switch i := i.(type) {
	case string:
		return i, nil
	}
	return "", errors.New("unknown string type")
}

func interfaceToBool(i interface{}) (bool, error) {
	switch i := i.(type) {
	case bool:
		return i, nil
	}
	if i, err := interfaceToInt(i); err == nil {
		return i != 0, nil
	}
	return false, errors.New("unknown boolean type")
}
func interfaceToBoolOrParse(i interface{}) (bool, error) {
	if i, err := interfaceToBool(i); err == nil {
		return i, nil
	}
	if s, err := interfaceToString(i); err == nil {
		return strconv.ParseBool(s)
	}
	return false, errors.New("unknown boolean/string type")
}
