package lookup

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type Reflector struct {
	path string
	v    reflect.Value
}

func (r Reflector) Type() reflect.Type {
	return r.v.Type()
}

func (r Reflector) Raw() interface{} {
	return r.v.Interface()
}

func (r Reflector) Value() reflect.Value {
	return r.v
}

func (r *Reflector) Find(path string, opts ...PathOpt) Pathor {
	settings := &PathSettings{}
	for _, opt := range opts {
		opt.PathOptSet(settings)
	}
	p := r.path
	rr := r.subPath(path, r.v, p, nil, settings)
	return rr
}

func (r *Reflector) subPath(path string, v reflect.Value, p string, pv *reflect.Value, settings *PathSettings) Pathor {
	var result Pathor
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("mil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = r.subPath(path, v.Elem(), p, &v, settings)
	case reflect.Interface:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("mil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = r.subPath(path, v.Elem(), p, nil, settings)
	case reflect.Array:
		result = r.ArrayOrSlicePath(p, path, v, settings)
	case reflect.Map:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("mil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = r.MapPath(p, path, v, settings)
	case reflect.Slice:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("mil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = r.ArrayOrSlicePath(p, path, v, settings)
	case reflect.Struct:
		if path == "" {
			return r
		}
		result = r.StructPath(p, path, v, pv, settings)
	case reflect.Func:
		pather := r.RunMethod(v, p)
		if pather != nil {
			result = pather.Find(path)
			break
		}
		result = &Invalidor{
			err:  fmt.Errorf("invalid element at simple path %s method call returned error %s", p, "invalid method"),
			path: p,
		}
	default:
		p += path
		result = &Invalidor{
			err:  fmt.Errorf("invalid element at simple path %s element was %s expected array,slice,map,struct,func", p, v.Kind()),
			path: p,
		}
	}
	if settings.Default != nil {
		if _, ok := result.(*Invalidor); ok {
			result = settings.Default
		}
	}
	return result
}

func (r Reflector) ArrayOrSliceForEachPath(prefix string, path string, v reflect.Value, settings *PathSettings) Pathor {
	typeCount := map[reflect.Type]int{}
	type Pair struct {
		Boxed   Pathor
		Unboxed Pathor
	}
	result := make([]*Pair, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		p := prefix + "[*]"
		vi := v.Index(i)
		vipath := &Pair{
			Boxed: (&Reflector{
				path: p,
				v:    vi,
			}).Find(path, settings.InferOps()...),
		}
		if _, ok := vipath.Boxed.(*Invalidor); ok {
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
	p := prefix + "[*]." + path
	switch len(typeCount) {
	case 0:
		return &Invalidor{
			err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because nothing matched query", "int", path),
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

func (r Reflector) ArrayOrSlicePath(prefix string, path string, v reflect.Value, settings *PathSettings) Pathor {
	p := prefix + "[" + strconv.Quote(path) + "]"
	i, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		if pather := r.ArrayOrSliceForEachPath(prefix, path, v, settings); pather != nil && pather != Pathor(nil) {
			return pather
		}
		return &Invalidor{
			err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because it couldn't become a integer %w and it wasn't a valid key", "int", path, err),
			path: p,
		}
	}
	l := int64(v.Len())
	if i < 0 {
		i += l
	}
	if i < 0 || i >= l {
		if pather := r.ArrayOrSliceForEachPath(prefix, path, v, settings); pather != nil {
			return pather
		}
		return &Invalidor{
			err:  fmt.Errorf("error looking up index of type %s value given was %#v and failed because %s", "int", path, "it was out of range"),
			path: p,
		}
	}
	rv := v.Index(int(i))
	return &Reflector{
		path: p,
		v:    rv,
	}
}

func (r Reflector) MapPath(prefix string, path string, v reflect.Value, settings *PathSettings) Pathor {
	p := prefix + "." + strconv.Quote(path)
	if prefix == "" || strings.HasSuffix(prefix, ".") {
		p = prefix + strconv.Quote(path)
	}
	//p := prefix + "[\"" + strconv.Quote(path) + "\"]"
	k, pather := r.ExtractKey(path, v, p)
	if pather != nil {
		return pather
	}
	ve := v.MapIndex(k)
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

func (r Reflector) ExtractKey(path string, v reflect.Value, p string) (reflect.Value, Pathor) {
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

func (r Reflector) StructPath(prefix string, path string, v reflect.Value, pv *reflect.Value, settings *PathSettings) Pathor {
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
	pather := r.RunMethod(m, p)
	if pather != nil {
		return pather
	}
	return &Invalidor{
		err:  fmt.Errorf("invalid element at simple path %s field or method was not found", p),
		path: p,
	}
}

func (r Reflector) RunMethod(m reflect.Value, p string) Pathor {
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

func Reflect(i interface{}) Pathor {
	return &Reflector{
		v: reflect.ValueOf(i),
	}
}
