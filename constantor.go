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

// PathOptSet allows for Constantor to be used as the default / fallback value on a .Find() operation
func (d *Constantor) PathOptSet(ctx *PathSettings) {
	ctx.Default = d
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
		switch opt := opt.(type) {
		case *Constantor:
			c = opt.c
		default:
			opt.PathOptSet(settings)
		}
	}

	nc := &Constantor{
		c:    c,
		path: p,
	}
	for _, evaluator := range settings.Evaluators {
		if e, err := evaluator.Evaluate(nc); err != nil {
			return NewInvalidor(p, err)
		} else if !e {
			return NewInvalidor(p, ErrEvalFail)
		}
	}

	return nc
}
