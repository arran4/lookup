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
