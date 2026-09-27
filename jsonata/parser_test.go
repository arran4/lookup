package jsonata

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestArrayConstructor(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"empty array", `[]`, false},
		{"single element", `[1]`, false},
		{"multiple elements", `[1, 2]`, false},
		{"nested array", `[[1], 2]`, false},
		{"arithmetic expression", `[1 + 2, 4]`, false},
		{"field paths", `[foo, bar]`, false},
		{"nested expression", `[[foo], bar]`, false},
		{"malformed comma", `[1,,2]`, true},
		{"malformed brackets", `[1,2`, true},
		{"trailing comma", `[1,]`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.expr)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ast)
			}
		})
	}
}

func TestParseObjectConstructor(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{"empty object", `{}`, false},
		{"single property string key", `{"key": "value"}`, false},
		{"single property ident key", `{key: "value"}`, false},
		{"multiple properties", `{"one": 1, "two": 2}`, false},
		{"nested object", `{"one": 1, "two": {"three": 3, "four": "4"}}`, false},
		{"array in object", `{"one": 1, "two": [3, "four"]}`, false},
		{"malformed comma", `{"one": 1,,}`, true},
		{"trailing comma", `{"one": 1,}`, true},
		{"missing colon", `{"one" 1}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := Parse(tt.expr)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ast)
			}
		})
	}
}
