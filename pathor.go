package lookup

import (
	"reflect"
)

type Finder interface {
	// Find preforms a path navigation. 'Path' is either a map key, array/slice index, or struct function/field.
	// This function is fixed and will probably not change in the future
	// So usage is supposed to be changed, anything which implements this function should return null-safe values. (ie non
	// nul.)
	// Usage: `lookup.Reflector(MyObjcet).Find("Quotes").Find("12").Find("Qty").Raw()
	Find(path string, opts ...Runner) Pathor
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
	// RawAsInterfaceSlice returns the raw contents as a slice of interface{}. If the value is not a slice it returns nil.
	RawAsInterfaceSlice() []interface{}
	// Type returns the reflect.Type or nil. This could be restricted to lookup.Reflector and others where appropriate
	Type() reflect.Type
}
