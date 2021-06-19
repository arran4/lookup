package lookup

import (
	"reflect"
)

// Invalidor indicates an invalid state this can be because of an error, or an invalid path. It contains an error and
// adheres to errors and errors.Unwrap using fmt.Errors(".. %w..") It is designed to be continued to be used without
// returning a null value when you reach an error and also provide the path and error combo for debugging. It is fully
// adherent to a Pathor object
type Invalidor struct {
	err  error
	path string
}

// NewInvalidor creates an invalidator, there shouldn't be any real reason to do this but you have an option to. See
// documentation for Invalidor for details
func NewInvalidor(path string, err error) *Invalidor {
	return &Invalidor{
		err:  err,
		path: path,
	}
}

// Type returns NULL
func (r Invalidor) Type() reflect.Type {
	return nil
}

// Raw returns NULL
func (r Invalidor) Raw() interface{} {
	return nil
}

// Raw returns a zero/invalid reflect.Value
func (r Invalidor) Value() reflect.Value {
	return reflect.Value{}
}

// Error implements the error interface
func (r Invalidor) Error() string {
	return r.err.Error()
}

// Unwrap implements the Unwrap error interface
func (r Invalidor) Unwrap() error {
	return r.err
}

// Find returns a new Invalidator with the same object but with an updated path if required. -- The path changing component
// might be removed - or become toggleable in an option.
func (r Invalidor) Find(path string, opts ...PathOpt) Pathor {
	p := r.path
	if len(r.path) > 0 {
		p = r.path + "." + path
	} else {
		p = path
	}
	return &Invalidor{
		err:  r.err,
		path: p,
	}
}
