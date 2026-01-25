package lookup

import (
	"encoding/json"
	"reflect"
)

// Jsonor is a Pathor that lazily unmarshals JSON bytes when accessed.
type Jsonor struct {
	path string
	raw  json.RawMessage
	p    Pathor
	done bool
}

// Json creates a Pathor for navigating raw JSON data.
func Json(raw []byte) Pathor {
	return &Jsonor{raw: json.RawMessage(raw)}
}

// Path returns the current lookup path.
func (j *Jsonor) Path() string { return j.path }

func (j *Jsonor) ensure() Pathor {
	if j.done {
		return j.p
	}
	j.done = true
	var v interface{}
	if err := json.Unmarshal(j.raw, &v); err != nil {
		j.p = NewInvalidor(j.path, err)
	} else {
		j.p = &Reflector{path: j.path, v: reflect.ValueOf(v)}
	}
	return j.p
}

// Find navigates the JSON structure using Reflector after decoding.
func (j *Jsonor) Find(path string, opts ...Runner) Pathor {
	return j.ensure().Find(path, opts...)
}

// Raw returns the decoded value.
func (j *Jsonor) Raw() interface{} { return j.ensure().Raw() }

// RawAsInterfaceSlice returns the decoded value as a slice of interface{}.
func (j *Jsonor) RawAsInterfaceSlice() []interface{} { return j.ensure().RawAsInterfaceSlice() }

// Value returns the reflect.Value of the decoded value.
func (j *Jsonor) Value() reflect.Value { return j.ensure().Value() }

// Type returns the reflect.Type of the decoded value.
func (j *Jsonor) Type() reflect.Type { return j.ensure().Type() }

func (j *Jsonor) IsString() bool    { return j.ensure().IsString() }
func (j *Jsonor) IsInt() bool       { return j.ensure().IsInt() }
func (j *Jsonor) IsBool() bool      { return j.ensure().IsBool() }
func (j *Jsonor) IsFloat() bool     { return j.ensure().IsFloat() }
func (j *Jsonor) IsSlice() bool     { return j.ensure().IsSlice() }
func (j *Jsonor) IsMap() bool       { return j.ensure().IsMap() }
func (j *Jsonor) IsStruct() bool    { return j.ensure().IsStruct() }
func (j *Jsonor) IsNil() bool       { return j.ensure().IsNil() }
func (j *Jsonor) IsPtr() bool       { return j.ensure().IsPtr() }
func (j *Jsonor) IsInterface() bool { return j.ensure().IsInterface() }

func (j *Jsonor) AsString() (string, error)              { return j.ensure().AsString() }
func (j *Jsonor) AsInt() (int64, error)                  { return j.ensure().AsInt() }
func (j *Jsonor) AsBool() (bool, error)                  { return j.ensure().AsBool() }
func (j *Jsonor) AsFloat() (float64, error)              { return j.ensure().AsFloat() }
func (j *Jsonor) AsSlice() ([]interface{}, error)        { return j.ensure().AsSlice() }
func (j *Jsonor) AsMap() (map[string]interface{}, error) { return j.ensure().AsMap() }
func (j *Jsonor) AsPtr() (interface{}, error)            { return j.ensure().AsPtr() }
