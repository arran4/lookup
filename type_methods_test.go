package lookup_test

import (
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

	// Test IsInt
	if !p.Find("int").IsInt() {
		t.Errorf("Expected IsInt() to be true for int value")
	}
	if p.Find("str").IsInt() {
		t.Errorf("Expected IsInt() to be false for string value")
	}

	// Test IsBool
	if !p.Find("bool").IsBool() {
		t.Errorf("Expected IsBool() to be true for bool value")
	}
	if p.Find("str").IsBool() {
		t.Errorf("Expected IsBool() to be false for string value")
	}

	// Test IsFloat
	if !p.Find("float").IsFloat() {
		t.Errorf("Expected IsFloat() to be true for float value")
	}
	if p.Find("int").IsFloat() {
		// Note: Int is not Float in reflect.Kind
		t.Errorf("Expected IsFloat() to be false for int value")
	}

	// Test IsSlice
	if !p.Find("slice").IsSlice() {
		t.Errorf("Expected IsSlice() to be true for slice value")
	}
	if p.Find("str").IsSlice() {
		t.Errorf("Expected IsSlice() to be false for string value")
	}

	// Test IsMap
	if !p.Find("map").IsMap() {
		t.Errorf("Expected IsMap() to be true for map value")
	}
	if p.Find("str").IsMap() {
		t.Errorf("Expected IsMap() to be false for string value")
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
}
