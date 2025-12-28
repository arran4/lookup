package lookup

import (
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
