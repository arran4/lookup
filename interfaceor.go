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

func (i *Interfaceor) Find(path string, opts ...Runner) Pathor {
	cp, _ := i.i.(CustomPath)
	p := PathBuilder(path, i, cp)
	var ni Pathor
	nii, err := i.i.Get(path)
	if err != nil {
		ni = NewInvalidor(p, err)
	} else {
		switch nii := nii.(type) {
		case Interface:
			ni = &Interfaceor{
				i:    nii,
				path: p,
			}
		case Pathor:
			ni = nii
		default:
			ni = &Invalidor{
				err:  fmt.Errorf("invalid return type: %s", reflect.TypeOf(ni)),
				path: p,
			}
		}
	}
	for _, evaluator := range opts {
		ni = evaluator.Run(NewScope(i, ni))
		if ni == nil {
			ni = NewInvalidor(p, ErrEvalFail)
		}
	}
	return ni
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
