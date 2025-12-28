package lookup

import (
	"fmt"
	"reflect"
)

// Reflector is a Pathor which uses reflection for navigation of the the objects, it supports a wide range of elements
type Reflector struct {
	path string
	v    reflect.Value
}

func (r *Reflector) Path() string {
	return r.path
}

// Type extracts the reflect.Type from the stored object
func (r *Reflector) Type() reflect.Type {
	return r.v.Type()
}

// Raw returns the contained object / reference.
func (r *Reflector) Raw() interface{} {
	if !r.v.IsValid() {
		return nil
	}
	return r.v.Interface()
}

// Value returns the reflect.Value
func (r *Reflector) Value() reflect.Value {
	return r.v
}

// Find finds the best match for the "Path" argument in the contained object and then returns a Pathor for that location
// Match nothing was found it will return an Invalidor, or if a Constant has bee provided as an argument (such as through
// `Default()` it will default to that in most cases. Find is designed to return null safe results.
func (r *Reflector) Find(path string, opts ...Runner) Pathor {
	rr := r.subPath(path, r.v, r.path, nil)
	p := ExtractPath(rr)
	for _, runner := range opts {
		rr = runner.Run(NewScope(r, rr))
		if rr == nil {
			rr = NewInvalidor(p, ErrEvalFail)
		}
	}
	return rr
}

// subPath determines type and preforms the correct action. -- Match an error defaults to default
func (r *Reflector) subPath(path string, v reflect.Value, p string, pv *reflect.Value) Pathor {
	if path == "" {
		return r
	}
	var result Pathor
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("nil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = r.subPath(path, v.Elem(), p, &v)
	case reflect.Interface:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("nil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = r.subPath(path, v.Elem(), p, nil)
	case reflect.Array:
		result = arrayOrSliceForEachPath(p, []string{path}, v, nil, nil)
	case reflect.Map:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("nil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = mapPath(p, path, v)
	case reflect.Slice:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("nil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = arrayOrSliceForEachPath(p, []string{path}, v, nil, nil)
	case reflect.Struct:
		if path == "" {
			result = r
			break
		}
		result = structPath(p, path, v, pv)
	case reflect.Func:
		pather := runMethod(v, p)
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
	return result
}

// Reflect creates a Pathor that uses reflect to navigate the object. This so far is the only way to navigate arbitrary
// go objects, so use this.
func Reflect(i interface{}) Pathor {
	if p, ok := i.(Pathor); ok {
		return p
	}
	return &Reflector{
		v: reflect.ValueOf(i),
	}
}

func (r *Reflector) IsString() bool {
	return r.v.Kind() == reflect.String
}

func (r *Reflector) IsInt() bool {
	k := r.v.Kind()
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64
}

func (r *Reflector) IsBool() bool {
	return r.v.Kind() == reflect.Bool
}

func (r *Reflector) IsFloat() bool {
	k := r.v.Kind()
	return k == reflect.Float32 || k == reflect.Float64
}

func (r *Reflector) IsSlice() bool {
	k := r.v.Kind()
	return k == reflect.Slice || k == reflect.Array
}

func (r *Reflector) IsMap() bool {
	return r.v.Kind() == reflect.Map
}

func (r *Reflector) IsStruct() bool {
	return r.v.Kind() == reflect.Struct
}

func (r *Reflector) IsNil() bool {
	if !r.v.IsValid() {
		return true
	}
	switch r.v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return r.v.IsNil()
	}
	return false
}

func (r *Reflector) IsPtr() bool {
	return r.v.Kind() == reflect.Ptr
}

func (r *Reflector) IsInterface() bool {
	return r.v.Kind() == reflect.Interface
}

func (r *Reflector) AsString() (string, error) {
	if r.IsString() {
		return r.v.String(), nil
	}
	return "", fmt.Errorf("path %s: %w", r.Path(), ErrNotString)
}

func (r *Reflector) AsInt() (int64, error) {
	if r.IsInt() {
		return r.v.Int(), nil
	}
	return 0, fmt.Errorf("path %s: %w", r.Path(), ErrNotInt)
}

func (r *Reflector) AsBool() (bool, error) {
	if r.IsBool() {
		return r.v.Bool(), nil
	}
	return false, fmt.Errorf("path %s: %w", r.Path(), ErrNotBool)
}

func (r *Reflector) AsFloat() (float64, error) {
	if r.IsFloat() {
		return r.v.Float(), nil
	}
	return 0.0, fmt.Errorf("path %s: %w", r.Path(), ErrNotFloat)
}

func (r *Reflector) AsSlice() ([]interface{}, error) {
	if r.IsSlice() {
		// Convert slice to []interface{}
		l := r.v.Len()
		res := make([]interface{}, l)
		for i := 0; i < l; i++ {
			res[i] = r.v.Index(i).Interface()
		}
		return res, nil
	}
	return nil, fmt.Errorf("path %s: %w", r.Path(), ErrNotSlice)
}

func (r *Reflector) AsMap() (map[string]interface{}, error) {
	if r.IsMap() {
		// Convert map to map[string]interface{} if keys are strings
		if r.v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("path %s: map keys are not strings", r.Path())
		}
		res := make(map[string]interface{})
		iter := r.v.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			res[k] = iter.Value().Interface()
		}
		return res, nil
	}
	return nil, fmt.Errorf("path %s: %w", r.Path(), ErrNotMap)
}

func (r *Reflector) AsPtr() (interface{}, error) {
	if r.IsPtr() {
		return r.v.Interface(), nil
	}
	return nil, fmt.Errorf("path %s: %w", r.Path(), ErrNotPtr)
}
