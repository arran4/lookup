package lookup

import (
	"reflect"

	"github.com/arran4/go-evaluator"
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
	Context  *evaluator.Context
}

func NewScope(parent Pathor, position Pathor) *Scope {
	return NewScopeWithContext(parent, position, nil)
}

func NewScopeWithContext(parent Pathor, position Pathor, ctx *evaluator.Context) *Scope {
	var parentScope *Scope
	if parent != nil {
		parentScope = NewScopeWithContext(nil, parent, ctx)
	}
	return &Scope{
		Current:  position,
		Parent:   parentScope,
		Position: position,
		Context:  ctx,
	}
}

func (s *Scope) Copy() *Scope {
	var ctx *evaluator.Context
	if s != nil {
		ctx = s.Context
	}
	return &Scope{
		Current: s.Current,
		Parent:  s.Parent,
		v:       s.v,
		path:    s.path,
		Context: ctx,
	}
}

func (s *Scope) Nest(new Pathor) *Scope {
	var ctx *evaluator.Context
	if s != nil {
		ctx = s.Context
	}
	return &Scope{
		Current:  new,
		Parent:   s,
		Position: new,
		Context:  ctx,
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
