package lookup

import (
	"reflect"
)

// PathSettings are the settings used by the Find function.
type PathSettings struct {
	Default    *Constantor
	Evaluators []*Evaluator
}

// InferOps reverse engineers the options provided to PathSettings
func (s *PathSettings) InferOps() []PathOpt {
	result := []PathOpt{}
	if s.Default != nil {
		result = append(result, s.Default)
	}
	for _, e := range s.Evaluators {
		result = append(result, e)
	}
	return result
}

// PathOpt an interface denoting what can be an option.
type PathOpt interface {
	PathOptSet(settings *PathSettings)
}

// NewDefault used with .Find() as a PathOpt this will will fallback / default to the provided value regardless of future
// nagivations, it suppresses most errors / Invalidators.
func NewDefault(i interface{}) PathOpt {
	return NewConstantor("", i)
}

type Finder interface {
	// Find preforms a path navigation. 'Path' is either a map key, array/slice index, or struct function/field.
	// This function is fixed and will probably not change in the future
	// So usage is supposed to be changed, anything which implements this function should return null-safe values. (ie non
	// nul.)
	// Usage: `lookup.Reflector(MyObjcet).Find("Quotes").Find("12").Find("Qty").Raw()
	Find(path string, opts ...PathOpt) Pathor
}

// Pathor interface
type Pathor interface {
	// Finder preforms a path navigation. 'Path' is either a map key, array/slice index, or struct function/field.
	// This function is fixed and will probably not change in the future
	// So usage is supposed to be changed, anything which implements this function should return null-safe values. (ie non
	// nul.)
	// Usage: `lookup.Reflector(MyObjcet).Find("Quotes").Find("12").Find("Qty").Raw()
	Finder
	// Value returns the reflect.Value or an invalid reflect.Value. This could be restricted to lookup.Reflector and others where appropriate
	Value() reflect.Value
	// Raw returns the raw contents / result of the lookup. This won't change
	Raw() interface{}
	// Type returns the reflect.Type or nil. This could be restricted to lookup.Reflector and others where appropriate
	Type() reflect.Type
}
