package lookup

import (
	"fmt"
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

func (c *Constantor) AsString() (string, error) {
	if s, ok := c.c.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("path %s: %w", c.path, ErrNotString)
}

func (c *Constantor) AsInt() (int64, error) {
	switch v := c.c.(type) {
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
	return 0, fmt.Errorf("path %s: %w", c.path, ErrNotInt)
}

func (c *Constantor) AsBool() (bool, error) {
	if v, ok := c.c.(bool); ok {
		return v, nil
	}
	return false, fmt.Errorf("path %s: %w", c.path, ErrNotBool)
}

func (c *Constantor) AsFloat() (float64, error) {
	switch v := c.c.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	}
	return 0.0, fmt.Errorf("path %s: %w", c.path, ErrNotFloat)
}

func (c *Constantor) AsSlice() ([]interface{}, error) {
	if c.IsSlice() {
		v := reflect.ValueOf(c.c)
		l := v.Len()
		res := make([]interface{}, l)
		for i := 0; i < l; i++ {
			res[i] = v.Index(i).Interface()
		}
		return res, nil
	}
	return nil, fmt.Errorf("path %s: %w", c.path, ErrNotSlice)
}

func (c *Constantor) AsMap() (map[string]interface{}, error) {
	if c.IsMap() {
		v := reflect.ValueOf(c.c)
		if v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("path %s: map keys are not strings", c.path)
		}
		res := make(map[string]interface{})
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			res[k] = iter.Value().Interface()
		}
		return res, nil
	}
	return nil, fmt.Errorf("path %s: %w", c.path, ErrNotMap)
}

func (c *Constantor) AsPtr() (interface{}, error) {
	if c.IsPtr() {
		return c.c, nil
	}
	return nil, fmt.Errorf("path %s: %w", c.path, ErrNotPtr)
}
