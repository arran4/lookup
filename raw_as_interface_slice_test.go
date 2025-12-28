package lookup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawAsInterfaceSlice_Reflector(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []interface{}
	}{
		{
			name:     "slice of ints",
			input:    []int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "array of strings",
			input:    [2]string{"a", "b"},
			expected: []interface{}{"a", "b"},
		},
		{
			name:     "not a slice",
			input:    "string",
			expected: nil,
		},
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Reflect(tt.input)
			assert.Equal(t, tt.expected, r.RawAsInterfaceSlice())
		})
	}
}

func TestRawAsInterfaceSlice_Simpleor(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []interface{}
	}{
		{
			name:     "slice of ints",
			input:    []int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "slice of interface",
			input:    []interface{}{1, "a", true},
			expected: []interface{}{1, "a", true},
		},
		{
			name:     "not a slice",
			input:    "string",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Simple(tt.input)
			assert.Equal(t, tt.expected, r.RawAsInterfaceSlice())
		})
	}
}

func TestRawAsInterfaceSlice_Constantor(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []interface{}
	}{
		{
			name:     "slice of ints",
			input:    []int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "not a slice",
			input:    "string",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Constant(tt.input)
			assert.Equal(t, tt.expected, r.RawAsInterfaceSlice())
		})
	}
}

func TestRawAsInterfaceSlice_Invalidor(t *testing.T) {
	r := NewInvalidor("path", nil)
	assert.Nil(t, r.RawAsInterfaceSlice())
}

func TestRawAsInterfaceSlice_Jsonor(t *testing.T) {
	jsonData := `[1, 2, 3]`
	r := Json([]byte(jsonData))
	expected := []interface{}{1.0, 2.0, 3.0} // JSON numbers are floats by default
	assert.Equal(t, expected, r.RawAsInterfaceSlice())

	jsonDataObj := `{"a": 1}`
	rObj := Json([]byte(jsonDataObj))
	assert.Nil(t, rObj.RawAsInterfaceSlice())
}

func TestRawAsInterfaceSlice_Yamlor(t *testing.T) {
	yamlData := `
- 1
- 2
- 3
`
	r := Yaml([]byte(yamlData))
	expected := []interface{}{1, 2, 3}
	assert.Equal(t, expected, r.RawAsInterfaceSlice())

	yamlDataObj := `a: 1`
	rObj := Yaml([]byte(yamlDataObj))
	assert.Nil(t, rObj.RawAsInterfaceSlice())
}
