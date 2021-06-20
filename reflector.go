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
	return r.v.Interface()
}

// Value returns the reflect.Value
func (r *Reflector) Value() reflect.Value {
	return r.v
}

// Evaluate returns not IsZero from reflect.Value for the purpose of implementing EvaluateNoArgs
func (r *Reflector) Evaluate() bool {
	if r.v.IsValid() {
		return !r.v.IsZero()
	}
	return false
}

// Find finds the best match for the "Path" argument in the contained object and then returns a Pathor for that location
// If nothing was found it will return an Invalidor, or if a Constant has bee provided as an argument (such as through
// `NewDefault()` it will default to that in most cases. Find is designed to return null safe results.
func (r *Reflector) Find(path string, opts ...PathOpt) Pathor {
	settings := &PathSettings{}
	for _, opt := range opts {
		opt.PathOptSet(settings)
	}
	rr := r.subPath(path, r.v, r.path, nil, settings)
	p := ExtractPath(rr)
	for _, evaluator := range settings.Evaluators {
		if e, err := evaluator.Evaluate(rr); err != nil {
			return NewInvalidor(p, err)
		} else if !e {
			return NewInvalidor(p, ErrEvalFail)
		}
	}

	return rr
}

// subPath determines type and preforms the correct action. -- If an error defautls to default
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
		result = arrayOrSlicePath(p, path, v, settings)
	case reflect.Map:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("mil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = mapPath(p, path, v, settings)
	case reflect.Slice:
		if v.IsNil() {
			p += path
			result = &Invalidor{
				err: fmt.Errorf("mil element at simple path %s element was %s expected array,slice,map,struct", p, "nil"),
			}
			break
		}
		result = arrayOrSlicePath(p, path, v, settings)
	case reflect.Struct:
		if path == "" {
			result = r
			break
		}
		result = structPath(p, path, v, pv, settings)
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
	if settings.Default != nil {
		if _, ok := result.(*Invalidor); ok {
			result = settings.Default
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
