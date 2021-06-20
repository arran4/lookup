package lookup

import (
	"reflect"
)

type Interface interface {
	Get(path string) (interface{}, error)
	Raw() interface{}
}

type Interfaceor struct {
	i    Interface
	path string
}

func (i *Interfaceor) Path() string {
	return i.path
}

func (i *Interfaceor) Find(path string, opts ...PathOpt) Pathor {
	cp, _ := i.i.(CustomPath)
	p := PathBuilder(path, i, cp)
	if ni, err := i.i.Get(path); err != nil {
		return NewInvalidor(p, err)
	} else if ni != nil {
		switch ni := ni.(type) {
		case Interface:
			return &Interfaceor{
				i:    ni,
				path: p,
			}
		}
	}
	settings := &PathSettings{}
	for _, opt := range opts {
		opt.PathOptSet(settings)
	}
	if settings.Default != nil {
		return settings.Default
	}
	return NewInvalidor(p, ErrNoSuchPath)
}

func (i *Interfaceor) Value() reflect.Value {
	return reflect.ValueOf(i.i.Raw())
}

func (i *Interfaceor) Raw() interface{} {
	return i.i.Raw()
}

func (i *Interfaceor) Type() reflect.Type {
	return reflect.TypeOf(i.i.Raw())
}

func NewInterfaceor(i Interface) Pathor {
	return &Interfaceor{
		i:    i,
		path: "",
	}
}
