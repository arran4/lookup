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
			return Reflect(s.v).Find(path, opts...)
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
