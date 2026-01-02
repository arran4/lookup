package jsonata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComments(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "Simple comment before identifier",
			expr: "/* comment */ foo",
		},
		{
			name: "Comment between identifiers",
			expr: "foo /* comment */ . bar",
		},
		{
			name: "Comment inside brackets",
			expr: "foo[ /* comment */ 0 /* comment */ ]",
		},
		{
			name: "Multiline comment",
			expr: "/* \n comment \n */ foo",
		},
		{
			name:    "Unclosed comment",
			expr:    "/* unclosed comment",
			wantErr: true,
			errMsg:  "unclosed comment",
		},
		{
			name:    "Unclosed comment with content",
			expr:    "foo /* unclosed",
			wantErr: true,
			errMsg:  "unclosed comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.expr)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
