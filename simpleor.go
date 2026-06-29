package lookup

import (
	"fmt"
	"reflect"
)

// Simpleor is a Pathor implementation that uses type switches for common types
// (map[string]interface{}, []interface{}) to avoid reflection overhead where possible.
type Simpleor struct {
	v    interface{}
	path string
}

func Simple(v interface{}) *Simpleor {
	return &Simpleor{v: v, path: ""}
}

func (s *Simpleor) Find(path string, opts ...Runner) Pathor {
	p := PathBuilder(path, s, nil)
	var nextV interface{}
	var err error

	if path == "" {
		nextV = s.v
	} else {
		switch tv := s.v.(type) {
		case map[string]interface{}:
			if val, ok := tv[path]; ok {
				nextV = val
			} else {
				err = ErrNoSuchPath
			}
		default:
			// Fallback to reflection for other types or deep navigation
			return (&Reflector{path: s.path, v: reflect.ValueOf(s.v)}).Find(path, opts...)
		}
	}

	var nextPathor Pathor
	if err != nil {
		nextPathor = NewInvalidor(p, err)
	} else {
		nextPathor = &Simpleor{v: nextV, path: p}
	}

	// Apply options
	// Scope.Parent is *Scope, not Pathor. We need to construct a parent scope if we want to support 'Parent' access correctly.
	// For now, let's create a new scope with the result.
	scope := NewScope(s, nextPathor)

	for _, opt := range opts {
		scope.Position = opt.Run(scope)
		if _, ok := scope.Position.(*Invalidor); ok {
			return scope.Position
		}
	}

	return scope.Position
}

func (s *Simpleor) Evaluate(scope *Scope, position Pathor) (Pathor, error) {
	return s, nil
}

func (s *Simpleor) Type() reflect.Type {
	return reflect.TypeOf(s.v)
}

func (s *Simpleor) Raw() interface{} {
	return s.v
}

func (s *Simpleor) RawAsInterfaceSlice() []interface{} {
	if s.v == nil {
		return nil
	}
	switch v := s.v.(type) {
	case []interface{}:
		return v
	}
	// Fallback to reflection if it's a different kind of slice?
	// Simpleor tries to avoid reflection.
	// But s.v can be anything.
	rv := reflect.ValueOf(s.v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		res := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			res[i] = rv.Index(i).Interface()
		}
		return res
	}
	return nil
}

func (s *Simpleor) Value() reflect.Value {
	return reflect.ValueOf(s.v)
}

func (s *Simpleor) Path() string {
	return s.path
}

func (s *Simpleor) IsString() bool {
	_, ok := s.v.(string)
	return ok
}

func (s *Simpleor) IsInt() bool {
	switch s.v.(type) {
	case int, int8, int16, int32, int64:
		return true
	}
	return false
}

func (s *Simpleor) IsBool() bool {
	_, ok := s.v.(bool)
	return ok
}

func (s *Simpleor) IsFloat() bool {
	switch s.v.(type) {
	case float32, float64:
		return true
	}
	return false
}

func (s *Simpleor) IsSlice() bool {
	if s.v == nil {
		return false
	}
	t := reflect.TypeOf(s.v)
	return t.Kind() == reflect.Slice || t.Kind() == reflect.Array
}

func (s *Simpleor) IsMap() bool {
	if s.v == nil {
		return false
	}
	return reflect.TypeOf(s.v).Kind() == reflect.Map
}

func (s *Simpleor) IsStruct() bool {
	if s.v == nil {
		return false
	}
	return reflect.TypeOf(s.v).Kind() == reflect.Struct
}

func (s *Simpleor) IsNil() bool {
	if s.v == nil {
		return true
	}
	// Also check if interface holds nil
	v := reflect.ValueOf(s.v)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func (s *Simpleor) IsPtr() bool {
	if s.v == nil {
		return false
	}
	return reflect.TypeOf(s.v).Kind() == reflect.Ptr
}

func (s *Simpleor) IsInterface() bool {
	// Everything in Simpleor is in interface{}, but we check underlying kind?
	// If checking if it is an interface, reflect.TypeOf(interface{}) just returns the type inside.
	// Unless s.v is actually an interface that hasn't been unwrapped?
	// But runtime types are never interface, only concrete types (or nil).
	// reflect.Kind() can be Interface only for reflect.Value of an interface field in a struct, not for interface{} value itself unless using elem?
	// Actually reflect.TypeOf(anyVar).Kind() will never be Interface.
	return false
}

func (s *Simpleor) AsString() (string, error) {
	if str, ok := s.v.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("path %s: %w", s.path, ErrNotString)
}

func (s *Simpleor) AsInt() (int64, error) {
	switch v := s.v.(type) {
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
	return 0, fmt.Errorf("path %s: %w", s.path, ErrNotInt)
}

func (s *Simpleor) AsBool() (bool, error) {
	if b, ok := s.v.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("path %s: %w", s.path, ErrNotBool)
}

func (s *Simpleor) AsFloat() (float64, error) {
	switch v := s.v.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	}
	return 0.0, fmt.Errorf("path %s: %w", s.path, ErrNotFloat)
}

func (s *Simpleor) AsSlice() ([]interface{}, error) {
	if s.IsSlice() {
		// Use reflection to convert slice to []interface{}
		v := reflect.ValueOf(s.v)
		l := v.Len()
		res := make([]interface{}, l)
		for i := 0; i < l; i++ {
			res[i] = v.Index(i).Interface()
		}
		return res, nil
	}
	return nil, fmt.Errorf("path %s: %w", s.path, ErrNotSlice)
}

func (s *Simpleor) AsMap() (map[string]interface{}, error) {
	if s.IsMap() {
		// If it's already map[string]interface{}, return it
		if m, ok := s.v.(map[string]interface{}); ok {
			return m, nil
		}
		// Otherwise use reflection
		v := reflect.ValueOf(s.v)
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("path %s: map keys are not strings", s.path)
		}
		res := make(map[string]interface{})
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			res[k] = iter.Value().Interface()
		}
		return res, nil
	}
	return nil, fmt.Errorf("path %s: %w", s.path, ErrNotMap)
}

func (s *Simpleor) AsPtr() (interface{}, error) {
	if s.IsPtr() {
		return s.v, nil
	}
	return nil, fmt.Errorf("path %s: %w", s.path, ErrNotPtr)
}
