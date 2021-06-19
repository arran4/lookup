package lookup

import (
	"reflect"
)

type Invalidor struct {
	err  error
	path string
}

func NewInvalidor(path string, err error) *Invalidor {
	return &Invalidor{
		err:  err,
		path: path,
	}
}

func (r Invalidor) Type() reflect.Type {
	return nil
}

func (r Invalidor) Raw() interface{} {
	return nil
}

func (r Invalidor) Value() reflect.Value {
	return reflect.Value{}
}

func (r Invalidor) Error() string {
	return r.err.Error()
}

func (r Invalidor) Unwrap() error {
	return r.err
}

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
