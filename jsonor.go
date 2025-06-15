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

// Value returns the reflect.Value of the decoded value.
func (j *Jsonor) Value() reflect.Value { return j.ensure().Value() }

// Type returns the reflect.Type of the decoded value.
func (j *Jsonor) Type() reflect.Type { return j.ensure().Type() }
