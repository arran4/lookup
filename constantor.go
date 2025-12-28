package lookup

import (
	"reflect"
)

// Constantor This object represents a non-navigable constant. It can be used as an argument applied on the appropriate
// location in a .Find() chain and it will be the fallback value if no value is found. It can be constructed with either
// lookup.NewConstantor or lookup.Default()
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

func Constant(c interface{}) *Constantor {
	return &Constantor{
		c: c,
	}
}

func True(path string) *Constantor {
	return &Constantor{
		c:    true,
		path: path,
	}
}

func False(path string) *Constantor {
	return &Constantor{
		c:    false,
		path: path,
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

// RawAsInterfaceSlice returns the contained object as a slice of interface{}.
func (r *Constantor) RawAsInterfaceSlice() []interface{} {
	if r.c == nil {
		return nil
	}
	rv := reflect.ValueOf(r.c)
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

// Value returns the reflect.Value
func (r *Constantor) Value() reflect.Value {
	return reflect.ValueOf(r.c)
}

// Find returns a new Constinator with the same object but with an updated path if required.
func (r *Constantor) Find(path string, opts ...Runner) Pathor {
	var p string
	if len(r.path) > 0 {
		p = r.path + "." + path
	} else {
		p = path
	}
	c := r.c
	var nc Pathor = &Constantor{
		c:    c,
		path: p,
	}
	for _, runner := range opts {
		nc = runner.Run(NewScope(r, nc))
		if nc == nil {
			nc = NewInvalidor(p, ErrEvalFail)
		}
	}

	return nc
}

func (c *Constantor) Run(scope *Scope) Pathor {
	return c
}
