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

func (c *Constantor) IsString() bool {
	_, ok := c.c.(string)
	return ok
}

func (c *Constantor) IsInt() bool {
	switch c.c.(type) {
	case int, int8, int16, int32, int64:
		return true
	}
	return false
}

func (c *Constantor) IsBool() bool {
	_, ok := c.c.(bool)
	return ok
}

func (c *Constantor) IsFloat() bool {
	switch c.c.(type) {
	case float32, float64:
		return true
	}
	return false
}

func (c *Constantor) IsSlice() bool {
	if c.c == nil {
		return false
	}
	t := reflect.TypeOf(c.c)
	return t.Kind() == reflect.Slice || t.Kind() == reflect.Array
}

func (c *Constantor) IsMap() bool {
	if c.c == nil {
		return false
	}
	return reflect.TypeOf(c.c).Kind() == reflect.Map
}

func (c *Constantor) IsStruct() bool {
	if c.c == nil {
		return false
	}
	return reflect.TypeOf(c.c).Kind() == reflect.Struct
}

func (c *Constantor) IsNil() bool {
	if c.c == nil {
		return true
	}
	v := reflect.ValueOf(c.c)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func (c *Constantor) IsPtr() bool {
	if c.c == nil {
		return false
	}
	return reflect.TypeOf(c.c).Kind() == reflect.Ptr
}

func (c *Constantor) IsInterface() bool {
	return false
}
