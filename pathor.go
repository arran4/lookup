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
	// Type returns the reflect.Type or nil. This could be restricted to lookup.Reflector and others where appropriate
	Type() reflect.Type

	// IsString returns true if the underlying value is a string
	IsString() bool
	// IsInt returns true if the underlying value is an int (int, int8, int16, int32, int64)
	IsInt() bool
	// IsBool returns true if the underlying value is a bool
	IsBool() bool
	// IsFloat returns true if the underlying value is a float (float32, float64)
	IsFloat() bool
	// IsSlice returns true if the underlying value is a slice or array
	IsSlice() bool
	// IsMap returns true if the underlying value is a map
	IsMap() bool
	// IsStruct returns true if the underlying value is a struct
	IsStruct() bool
	// IsNil returns true if the underlying value is nil
	IsNil() bool
	// IsPtr returns true if the underlying value is a pointer
	IsPtr() bool
	// IsInterface returns true if the underlying value is an interface
	IsInterface() bool
}
