package lookup

import (
	"reflect"

	"gopkg.in/yaml.v3"
)

// Yamlor is a Pathor that lazily unmarshals YAML bytes when accessed.
type Yamlor struct {
	path string
	raw  []byte
	p    Pathor
	done bool
}

// Yaml creates a Pathor for navigating raw YAML data.
func Yaml(raw []byte) Pathor {
	return &Yamlor{raw: raw}
}

// Path returns the current lookup path.
func (y *Yamlor) Path() string { return y.path }

func (y *Yamlor) ensure() Pathor {
	if y.done {
		return y.p
	}
	y.done = true
	var v interface{}
	if err := yaml.Unmarshal(y.raw, &v); err != nil {
		y.p = NewInvalidor(y.path, err)
	} else {
		y.p = &Reflector{path: y.path, v: reflect.ValueOf(v)}
	}
	return y.p
}

// Find navigates the YAML structure using Reflector after decoding.
func (y *Yamlor) Find(path string, opts ...Runner) Pathor {
	return y.ensure().Find(path, opts...)
}

// Raw returns the decoded value.
func (y *Yamlor) Raw() interface{} { return y.ensure().Raw() }

// RawAsInterfaceSlice returns the decoded value as a slice of interface{}.
func (y *Yamlor) RawAsInterfaceSlice() []interface{} { return y.ensure().RawAsInterfaceSlice() }

// Value returns the reflect.Value of the decoded value.
func (y *Yamlor) Value() reflect.Value { return y.ensure().Value() }

// Type returns the reflect.Type of the decoded value.
func (y *Yamlor) Type() reflect.Type { return y.ensure().Type() }

func (y *Yamlor) IsString() bool    { return y.ensure().IsString() }
func (y *Yamlor) IsInt() bool       { return y.ensure().IsInt() }
func (y *Yamlor) IsBool() bool      { return y.ensure().IsBool() }
func (y *Yamlor) IsFloat() bool     { return y.ensure().IsFloat() }
func (y *Yamlor) IsSlice() bool     { return y.ensure().IsSlice() }
func (y *Yamlor) IsMap() bool       { return y.ensure().IsMap() }
func (y *Yamlor) IsStruct() bool    { return y.ensure().IsStruct() }
func (y *Yamlor) IsNil() bool       { return y.ensure().IsNil() }
func (y *Yamlor) IsPtr() bool       { return y.ensure().IsPtr() }
func (y *Yamlor) IsInterface() bool { return y.ensure().IsInterface() }

func (y *Yamlor) AsString() (string, error)              { return y.ensure().AsString() }
func (y *Yamlor) AsInt() (int64, error)                  { return y.ensure().AsInt() }
func (y *Yamlor) AsBool() (bool, error)                  { return y.ensure().AsBool() }
func (y *Yamlor) AsFloat() (float64, error)              { return y.ensure().AsFloat() }
func (y *Yamlor) AsSlice() ([]interface{}, error)        { return y.ensure().AsSlice() }
func (y *Yamlor) AsMap() (map[string]interface{}, error) { return y.ensure().AsMap() }
func (y *Yamlor) AsPtr() (interface{}, error)            { return y.ensure().AsPtr() }
