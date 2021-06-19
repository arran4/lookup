package lookup

import (
	"reflect"
)

type PathSettings struct {
	Default *Constantor
}

func (s *PathSettings) InferOps() []PathOpt {
	result := []PathOpt{}
	if s.Default != nil {
		result = append(result, s.Default)
	}
	return result
}

type PathOpt interface {
	PathOptSet(settings *PathSettings)
}

func NewDefault(i interface{}) PathOpt {
	return NewConstantor("", i)
}

type Pathor interface {
	Find(path string, opts ...PathOpt) Pathor
	Value() reflect.Value
	Raw() interface{}
	Type() reflect.Type
}
