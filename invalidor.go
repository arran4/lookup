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

func (i *Invalidor) Path() string {
	return i.path
}

// Evaluate implements EvaluateNoArgs
func (i *Invalidor) Evaluate(scope *Scope, position Pathor) (Pathor, error) {
	return nil, nil
}

// Type returns NULL
func (i *Invalidor) Type() reflect.Type {
	return nil
}

// Raw returns NULL
func (i *Invalidor) Raw() interface{} {
	return nil
}

// RawAsInterfaceSlice returns nil
func (i *Invalidor) RawAsInterfaceSlice() []interface{} {
	return nil
}

// Raw returns a zero/invalid reflect.Value
func (i *Invalidor) Value() reflect.Value {
	return reflect.Value{}
}

// Error implements the error interface
func (i *Invalidor) Error() string {
	return i.err.Error()
}

// Unwrap implements the Unwrap error interface
func (i *Invalidor) Unwrap() error {
	return i.err
}

// Find returns a new Invalidator with the same object but with an updated path if required. -- The path changing component
// might be removed - or become toggleable in an option.
func (i *Invalidor) Find(path string, opts ...Runner) Pathor {
	p := PathBuilder(path, i, nil)
	return &Invalidor{
		err:  i.err,
		path: p,
	}
}

type errorFunc struct {
	err error
}

func Error(err error) *errorFunc {
	return &errorFunc{err: err}
}

func (e *errorFunc) Run(scope *Scope) Pathor {
	return NewInvalidor(scope.Path(), e.err)
}
