package lookup

import (
	"fmt"
	"reflect"
)

// Interface an interface you can implement to avoid using Reflector or to put your own selection logic such as if you
// were to run this over another data structure.
type Interface interface {
	// Find the next component.. Must return an Interface OR another type of Pathor.
	Get(path string) (interface{}, error)
	// The raw type
	Raw() interface{}
}

// Interfaceor the warping element for the Interface component to make it adhere to the Pathor interface
type Interfaceor struct {
	i    Interface
	path string
}

func (i *Interfaceor) Path() string {
	return i.path
}

func (i *Interfaceor) Find(path string, opts ...PathOpt) Pathor {
	cp, _ := i.i.(CustomPath)
	settings := &PathSettings{}
	for _, opt := range opts {
		opt.PathOptSet(settings)
	}
	p := PathBuilder(path, i, cp)
	finalError := NewInvalidor(p, ErrNoSuchPath)
	if ni, err := i.i.Get(path); err != nil {
		return NewInvalidor(p, err)
	} else if ni != nil {
		var np Pathor
		switch ni := ni.(type) {
		case Interface:
			np = &Interfaceor{
				i:    ni,
				path: p,
			}
		case Pathor:
			np = ni
		default:
			return &Invalidor{
				err:  fmt.Errorf("invalid return type: %s", reflect.TypeOf(ni)),
				path: p,
			}
		}
		pass := true
		finalError = NewInvalidor(p, ErrEvalFail)
		for _, evaluator := range settings.Evaluators {
			scope := &Scope{
				Current: np,
			}
			if e, err := evaluator.Evaluate(scope, np); err != nil {
				pass = false
				finalError = NewInvalidor(p, err)
				break
			} else if e != nil {
				np = e
			} else if e == nil {
				pass = false
				break
			}
		}
		if pass {
			return np
		}
	}
	if settings.Default != nil {
		return settings.Default
	}
	return finalError
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

// NewInterfaceor see Interface and Interfaceor for details.
func NewInterfaceor(i Interface) Pathor {
	return &Interfaceor{
		i:    i,
		path: "",
	}
}
