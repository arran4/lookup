package jsonata

import (
	"errors"
	"github.com/arran4/lookup"
	"reflect"
	"strings"
)

// Undefined represents the JSONata concept of no value.
type Undefined struct{}

// Sequence represents a JSONata sequence of values.
type Sequence struct {
	Values []interface{}
}

// Array represents a JSON array value. It preserves its array identity
// and is not sequence-flattened simply because it contains one item or nested arrays.
type Array struct {
	Elements []interface{}
}

// Sequence Flattening: Nested sequences are flattened into the parent sequence.
// Nested arrays are NOT flattened just because they are sequences.
func FlattenSequence(items ...interface{}) *Sequence {
	var flat []interface{}
	for _, item := range items {
		if seq, ok := item.(*Sequence); ok {
			// recursively flatten
			flat = append(flat, FlattenSequence(seq.Values...).Values...)
		} else if _, ok := item.(Undefined); ok {
			continue // undefined is empty sequence/omitted
		} else if item == nil {
			flat = append(flat, nil) // explicit null
		} else {
			flat = append(flat, item)
		}
	}
	return &Sequence{Values: flat}
}

// Materialize processes a raw result from execution into a standard JSON-compatible
// structure (or Undefined) according to JSONata rules.
// This unwraps Sequences according to singleton rules, and converts Array structs to slices.
func Materialize(val interface{}) interface{} {
	if val == nil {
		return nil // JSON null
	}

	switch v := val.(type) {
	case Undefined:
		return v
	case *Sequence:
		if len(v.Values) == 0 {
			return Undefined{}
		}
		if len(v.Values) == 1 {
			// Singleton sequence unwraps
			return Materialize(v.Values[0])
		}
		// Multi-value sequence materializes as array
		res := make([]interface{}, len(v.Values))
		for i, item := range v.Values {
			res[i] = Materialize(item)
		}
		return res
	case *Array:
		res := make([]interface{}, len(v.Elements))
		for i, item := range v.Elements {
			res[i] = Materialize(item)
		}
		return res
	case []interface{}:
		res := make([]interface{}, len(v))
		for i, item := range v {
			res[i] = Materialize(item)
		}
		return res
	case map[string]interface{}:
		res := make(map[string]interface{}, len(v))
		for k, item := range v {
			res[k] = Materialize(item)
		}
		return res
	default:
		return v
	}
}

// JSONata Truthiness semantics
// https://docs.jsonata.org/predicate#truthy-and-falsy-values
func Truthy(val interface{}) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case Undefined:
		return false
	case *Sequence:
		if len(v.Values) == 0 {
			return false
		}
		if len(v.Values) == 1 {
			return Truthy(v.Values[0])
		}
		for _, item := range v.Values {
			if Truthy(item) {
				return true
			}
		}
		return false
	case *Array:
		if len(v.Elements) == 0 {
			return false
		}
		for _, item := range v.Elements {
			if Truthy(item) {
				return true
			}
		}
		return false
	case []interface{}:
		if len(v) == 0 {
			return false
		}
		for _, item := range v {
			if Truthy(item) {
				return true
			}
		}
		return false
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		return len(v) > 0
	}

	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if rv.Len() == 0 {
			return false
		}
		for i := 0; i < rv.Len(); i++ {
			if Truthy(rv.Index(i).Interface()) {
				return true
			}
		}
		return false
	case reflect.Map:
		return rv.Len() > 0
	}

	return true
}

func IsUndefinedError(inv *lookup.Invalidor) bool {
	if inv == nil {
		return false
	}
	err := inv.Unwrap()
	if errors.Is(err, lookup.ErrNoSuchPath) {
		return true
	}
	errStr := inv.Error()
	return strings.Contains(errStr, "element not found at simple path")
}
