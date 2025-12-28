package lookup

import (
	"fmt"
	"reflect"
)

// Interface an interface you can implement to avoid using Reflector or to put your own selection logic such as if you
// were to run this over another data structure.
type Interface interface {
	// Find the next component.. Must return an Interface OR another type of Pathor.
	Get(path string) (interface{}, error)
	// The raw type
	Raw() interface{}
}

// Interfaceor the wrapping element for the Interface component to make it adhere to the Pathor interface
type Interfaceor struct {
	i    Interface
	path string
}

func (i *Interfaceor) Path() string {
	return i.path
}

func (i *Interfaceor) Find(path string, opts ...Runner) Pathor {
	cp, _ := i.i.(CustomPath)
	p := PathBuilder(path, i, cp)
	var ni Pathor
	nii, err := i.i.Get(path)
	if err != nil {
		ni = NewInvalidor(p, err)
	} else {
		switch nii := nii.(type) {
		case Interface:
			ni = &Interfaceor{
				i:    nii,
				path: p,
			}
		case Pathor:
			ni = nii
		default:
			ni = &Invalidor{
				err:  fmt.Errorf("invalid return type: %s", reflect.TypeOf(nii)),
				path: p,
			}
		}
	}
	for _, evaluator := range opts {
		ni = evaluator.Run(NewScope(i, ni))
		if ni == nil {
			ni = NewInvalidor(p, ErrEvalFail)
		}
	}
	return ni
}

func (i *Interfaceor) Value() reflect.Value {
	return reflect.ValueOf(i.i.Raw())
}

func (i *Interfaceor) Raw() interface{} {
	return i.i.Raw()
}

func (i *Interfaceor) Type() reflect.Type {
	return reflect.TypeOf(i.i.Raw())
}

func (i *Interfaceor) IsString() bool {
	_, ok := i.i.Raw().(string)
	return ok
}

func (i *Interfaceor) IsInt() bool {
	switch i.i.Raw().(type) {
	case int, int8, int16, int32, int64:
		return true
	}
	return false
}

func (i *Interfaceor) IsBool() bool {
	_, ok := i.i.Raw().(bool)
	return ok
}

func (i *Interfaceor) IsFloat() bool {
	switch i.i.Raw().(type) {
	case float32, float64:
		return true
	}
	return false
}

func (i *Interfaceor) IsSlice() bool {
	raw := i.i.Raw()
	if raw == nil {
		return false
	}
	t := reflect.TypeOf(raw)
	return t.Kind() == reflect.Slice || t.Kind() == reflect.Array
}

func (i *Interfaceor) IsMap() bool {
	raw := i.i.Raw()
	if raw == nil {
		return false
	}
	return reflect.TypeOf(raw).Kind() == reflect.Map
}

func (i *Interfaceor) IsStruct() bool {
	raw := i.i.Raw()
	if raw == nil {
		return false
	}
	return reflect.TypeOf(raw).Kind() == reflect.Struct
}

func (i *Interfaceor) IsNil() bool {
	raw := i.i.Raw()
	if raw == nil {
		return true
	}
	v := reflect.ValueOf(raw)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func (i *Interfaceor) IsPtr() bool {
	raw := i.i.Raw()
	if raw == nil {
		return false
	}
	return reflect.TypeOf(raw).Kind() == reflect.Ptr
}

func (i *Interfaceor) IsInterface() bool {
	return false
}

func (i *Interfaceor) AsString() (string, error) {
	if s, ok := i.i.Raw().(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("path %s: %w", i.path, ErrNotString)
}

func (i *Interfaceor) AsInt() (int64, error) {
	switch v := i.i.Raw().(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	}
	return 0, fmt.Errorf("path %s: %w", i.path, ErrNotInt)
}

func (i *Interfaceor) AsBool() (bool, error) {
	if v, ok := i.i.Raw().(bool); ok {
		return v, nil
	}
	return false, fmt.Errorf("path %s: %w", i.path, ErrNotBool)
}

func (i *Interfaceor) AsFloat() (float64, error) {
	switch v := i.i.Raw().(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	}
	return 0.0, fmt.Errorf("path %s: %w", i.path, ErrNotFloat)
}

func (i *Interfaceor) AsSlice() ([]interface{}, error) {
	if i.IsSlice() {
		v := reflect.ValueOf(i.i.Raw())
		l := v.Len()
		res := make([]interface{}, l)
		for idx := 0; idx < l; idx++ {
			res[idx] = v.Index(idx).Interface()
		}
		return res, nil
	}
	return nil, fmt.Errorf("path %s: %w", i.path, ErrNotSlice)
}

func (i *Interfaceor) AsMap() (map[string]interface{}, error) {
	if i.IsMap() {
		v := reflect.ValueOf(i.i.Raw())
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("path %s: map keys are not strings", i.path)
		}
		res := make(map[string]interface{})
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			res[k] = iter.Value().Interface()
		}
		return res, nil
	}
	return nil, fmt.Errorf("path %s: %w", i.path, ErrNotMap)
}

func (i *Interfaceor) AsPtr() (interface{}, error) {
	if i.IsPtr() {
		return i.i.Raw(), nil
	}
	return nil, fmt.Errorf("path %s: %w", i.path, ErrNotPtr)
}

// NewInterfaceor see Interface and Interfaceor for details.
func NewInterfaceor(i Interface) Pathor {
	return &Interfaceor{
		i:    i,
		path: "",
	}
}
