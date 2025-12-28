package lookup_test

import (
	"errors"
	"testing"

	"github.com/arran4/lookup"
)

func TestPathorTypeMethods(t *testing.T) {
	// Setup test data
	data := map[string]interface{}{
		"str":   "hello",
		"int":   123,
		"bool":  true,
		"float": 123.456,
		"slice": []interface{}{1, 2, 3},
		"map":   map[string]interface{}{"a": 1},
		"nil":   nil,
	}

	p := lookup.Reflect(data)

	// Test IsString
	if !p.Find("str").IsString() {
		t.Errorf("Expected IsString() to be true for string value")
	}
	if p.Find("int").IsString() {
		t.Errorf("Expected IsString() to be false for int value")
	}

	// Test AsString
	if s, err := p.Find("str").AsString(); err != nil || s != "hello" {
		t.Errorf("Expected AsString() to return 'hello', got '%s', err: %v", s, err)
	}
	if _, err := p.Find("int").AsString(); err == nil || !errors.Is(err, lookup.ErrNotString) {
		t.Errorf("Expected AsString() to return error for int value: %v", err)
	}

	// Test IsInt
	if !p.Find("int").IsInt() {
		t.Errorf("Expected IsInt() to be true for int value")
	}
	if p.Find("str").IsInt() {
		t.Errorf("Expected IsInt() to be false for string value")
	}

	// Test AsInt
	if i, err := p.Find("int").AsInt(); err != nil || i != 123 {
		t.Errorf("Expected AsInt() to return 123, got %d, err: %v", i, err)
	}
	if _, err := p.Find("str").AsInt(); err == nil || !errors.Is(err, lookup.ErrNotInt) {
		t.Errorf("Expected AsInt() to return error for string value: %v", err)
	}

	// Test IsBool
	if !p.Find("bool").IsBool() {
		t.Errorf("Expected IsBool() to be true for bool value")
	}
	if p.Find("str").IsBool() {
		t.Errorf("Expected IsBool() to be false for string value")
	}

	// Test AsBool
	if b, err := p.Find("bool").AsBool(); err != nil || !b {
		t.Errorf("Expected AsBool() to return true, got %v, err: %v", b, err)
	}
	if _, err := p.Find("str").AsBool(); err == nil || !errors.Is(err, lookup.ErrNotBool) {
		t.Errorf("Expected AsBool() to return error for string value: %v", err)
	}

	// Test IsFloat
	if !p.Find("float").IsFloat() {
		t.Errorf("Expected IsFloat() to be true for float value")
	}
	if p.Find("int").IsFloat() {
		// Note: Int is not Float in reflect.Kind
		t.Errorf("Expected IsFloat() to be false for int value")
	}

	// Test AsFloat
	if f, err := p.Find("float").AsFloat(); err != nil || f != 123.456 {
		t.Errorf("Expected AsFloat() to return 123.456, got %f, err: %v", f, err)
	}
	if _, err := p.Find("str").AsFloat(); err == nil || !errors.Is(err, lookup.ErrNotFloat) {
		t.Errorf("Expected AsFloat() to return error for string value: %v", err)
	}

	// Test IsSlice
	if !p.Find("slice").IsSlice() {
		t.Errorf("Expected IsSlice() to be true for slice value")
	}
	if p.Find("str").IsSlice() {
		t.Errorf("Expected IsSlice() to be false for string value")
	}

	// Test AsSlice
	if sl, err := p.Find("slice").AsSlice(); err != nil || len(sl) != 3 {
		t.Errorf("Expected AsSlice() to return slice of length 3, got %v, err: %v", sl, err)
	}
	if _, err := p.Find("str").AsSlice(); err == nil || !errors.Is(err, lookup.ErrNotSlice) {
		t.Errorf("Expected AsSlice() to return error for string value: %v", err)
	}

	// Test IsMap
	if !p.Find("map").IsMap() {
		t.Errorf("Expected IsMap() to be true for map value")
	}
	if p.Find("str").IsMap() {
		t.Errorf("Expected IsMap() to be false for string value")
	}

	// Test AsMap
	if m, err := p.Find("map").AsMap(); err != nil || m["a"] != 1 {
		t.Errorf("Expected AsMap() to return map with a=1, got %v, err: %v", m, err)
	}
	if _, err := p.Find("str").AsMap(); err == nil || !errors.Is(err, lookup.ErrNotMap) {
		t.Errorf("Expected AsMap() to return error for string value: %v", err)
	}

	// Test IsNil
	// p.Find("nil") should return a value that IsNil() returns true for.
	if !p.Find("nil").IsNil() {
		t.Errorf("Expected IsNil() to be true for nil value")
	}
	if p.Find("str").IsNil() {
		t.Errorf("Expected IsNil() to be false for string value")
	}

	// Test IsStruct
	type MyStruct struct {
		Field string
	}
	structData := MyStruct{Field: "value"}
	pStruct := lookup.Reflect(structData)

	if !pStruct.IsStruct() {
		t.Errorf("Expected IsStruct() to be true for struct value")
	}
	if pStruct.Find("Field").IsStruct() {
		t.Errorf("Expected IsStruct() to be false for string field")
	}

	// Test IsPtr
	ptrVal := &structData
	pPtr := lookup.Reflect(ptrVal)
	if !pPtr.IsPtr() {
		t.Errorf("Expected IsPtr() to be true for pointer value")
	}
	if pStruct.IsPtr() {
		t.Errorf("Expected IsPtr() to be false for struct value")
	}

	// Test AsPtr
	if ptr, err := pPtr.AsPtr(); err != nil || ptr != ptrVal {
		t.Errorf("Expected AsPtr() to return correct pointer, got %v, err: %v", ptr, err)
	}
	if _, err := pStruct.AsPtr(); err == nil || !errors.Is(err, lookup.ErrNotPtr) {
		t.Errorf("Expected AsPtr() to return error for struct value: %v", err)
	}
}
