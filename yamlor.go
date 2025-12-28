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
