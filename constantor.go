package lookup

import (
	"reflect"
)

// Constantor This object represents a non-navigable constant. It can be used as an argument applied on the appropriate
// location in a .Find() chain and it will be the fallback value if no value is found. It can be constructed with either
// lookup.NewConstantor or lookup.NewDefault()
type Constantor struct {
	path string
	c    interface{}
}

// NewConstantor constructs a non-navigable constant.
func NewConstantor(path string, c interface{}) *Constantor {
	return &Constantor{
		path: path,
		c:    c,
	}
}

// Constant constructs a non-navigable constant.
func Constant(c interface{}) *Constantor {
	return &Constantor{
		c: c,
	}
}

func Array(c ...interface{}) *Constantor {
	return &Constantor{
		path: "",
		c:    c,
	}
}

// Type extracts the reflect.Type from the stored object
func (r *Constantor) Type() reflect.Type {
	return reflect.TypeOf(r.c)
}

// Raw returns the contained object / reference.
func (r *Constantor) Raw() interface{} {
	return r.c
}

// Value returns the reflect.Value
func (r *Constantor) Value() reflect.Value {
	return reflect.ValueOf(r.c)
}

// Find returns a new Constinator with the same object but with an updated path if required.
func (r *Constantor) Find(path string, opts ...PathOpt) Pathor {
	p := r.path
	if len(r.path) > 0 {
		p = r.path + "." + path
	} else {
		p = path
	}
	c := r.c
	settings := &PathSettings{}
	for _, opt := range opts {
		opt.PathOptSet(settings)
	}

	var nc Pathor = &Constantor{
		c:    c,
		path: p,
	}
	for _, evaluator := range settings.Evaluators {
		scope := &Scope{
			Current: nc,
			Parent:  settings.Scope,
		}
		nc = evaluator.Evaluate(scope, nc)
		if nc == nil {
			nc = NewInvalidor(p, ErrEvalFail)
		}
	}

	return nc
}

func (c *Constantor) Run(scope *Scope, position Pathor) Pathor {
	return c
}
