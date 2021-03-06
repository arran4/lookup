package lookup

import (
	"reflect"
)

type Runner interface {
	Run(scope *Scope) Pathor
}

type Scope struct {
	Current  Pathor
	Parent   *Scope
	v        reflect.Value
	path     *string
	Position Pathor
}

func NewScope(parent Pathor, position Pathor) *Scope {
	var parentScope *Scope
	if parent != nil {
		parentScope = NewScope(nil, parent)
	}
	return &Scope{
		Current:  position,
		Parent:   parentScope,
		Position: position,
	}
}

func (s *Scope) Copy() *Scope {
	return &Scope{
		Current: s.Current,
		Parent:  s.Parent,
		v:       s.v,
		path:    s.path,
	}
}

func (s *Scope) Nest(new Pathor) *Scope {
	return &Scope{
		Current:  new,
		Parent:   s,
		Position: new,
	}
}

func (s *Scope) Value() reflect.Value {
	if s.v.IsValid() {
		return s.v
	}
	if s.Current != nil {
		return s.Current.Value()
	}
	if s.Parent != nil {
		return s.Parent.Value()
	}
	return reflect.Value{}
}

func (s *Scope) Path() string {
	if s.path != nil {
		return *s.path
	}
	if s.Parent != nil {
		return s.Parent.Path()
	}
	return ExtractPath(s.Current)
}

func (s *Scope) Next(position Pathor) *Scope {
	return &Scope{
		Current:  s.Current,
		Parent:   s,
		Position: position,
	}
}
