package lookup

import (
	"reflect"
)

type Constantor struct {
	path string
	c    interface{}
}

func NewConstantor(path string, c interface{}) *Constantor {
	return &Constantor{
		path: path,
		c:    c,
	}
}

func (r *Constantor) Type() reflect.Type {
	return reflect.TypeOf(r.c)
}

func (r *Constantor) Raw() interface{} {
	return r.c
}

func (r *Constantor) Value() reflect.Value {
	return reflect.ValueOf(r.c)
}

func (d *Constantor) PathOptSet(ctx *PathSettings) {
	ctx.Default = d
}

func (r *Constantor) Find(path string, opts ...PathOpt) Pathor {
	p := r.path
	if len(r.path) > 0 {
		p = r.path + "." + path
	} else {
		p = path
	}
	c := r.c
	for _, opt := range opts {
		switch opt := opt.(type) {
		case *Constantor:
			c = opt.c
		}
	}
	return &Constantor{
		c:    c,
		path: p,
	}
}
