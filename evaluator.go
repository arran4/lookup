package lookup

import (
	"reflect"
)

type Runner interface {
	Run(scope *Scope, position Pathor) Pathor
}

type Scope struct {
	Current Pathor
	Parent  *Scope
	Args    []Pathor
	v       reflect.Value
	path    *string
}

func (s *Scope) Copy() *Scope {
	return &Scope{
		Current: s.Current,
		Parent:  s.Parent,
		Args:    append([]Pathor{}, s.Args...),
		v:       s.v,
		path:    s.path,
	}
}

func (s *Scope) Nest(new Pathor, args ...Pathor) *Scope {
	return &Scope{
		Current: new,
		Parent:  s,
		Args:    args,
	}
}

func (s *Scope) Value() reflect.Value {
	if s.v.IsValid() {
		return s.v
	}
	if s.Parent != nil {
		return s.Parent.Value()
	}
	return s.Current.Value()
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
