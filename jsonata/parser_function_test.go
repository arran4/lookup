package jsonata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected *AST
		wantErr  bool
	}{
		{
			name: "Simple function call no args",
			expr: "foo()",
			expected: &AST{
				Node: &PathNode{
					Steps: []Step{
						{
							FunctionCall: &FunctionCallNode{
								Name: "foo",
								Args: nil,
							},
						},
					},
				},
			},
		},
		{
			name: "Function call one arg",
			expr: "foo(1)",
			expected: &AST{
				Node: &PathNode{
					Steps: []Step{
						{
							FunctionCall: &FunctionCallNode{
								Name: "foo",
								Args: []Node{
									&LiteralNode{Value: float64(1)},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Function call multiple args",
			expr: "foo(1, 'bar')",
			expected: &AST{
				Node: &PathNode{
					Steps: []Step{
						{
							FunctionCall: &FunctionCallNode{
								Name: "foo",
								Args: []Node{
									&LiteralNode{Value: float64(1)},
									&LiteralNode{Value: "bar"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Nested function calls",
			expr: "foo(bar(1))",
			expected: &AST{
				Node: &PathNode{
					Steps: []Step{
						{
							FunctionCall: &FunctionCallNode{
								Name: "foo",
								Args: []Node{
									&PathNode{
										Steps: []Step{
											{
												FunctionCall: &FunctionCallNode{
													Name: "bar",
													Args: []Node{
														&LiteralNode{Value: float64(1)},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Function call in path",
			expr: "foo.bar(1)",
			expected: &AST{
				Node: &PathNode{
					Steps: []Step{
						{Name: "foo"},
						{
							FunctionCall: &FunctionCallNode{
								Name: "bar",
								Args: []Node{
									&LiteralNode{Value: float64(1)},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Unclosed parenthesis",
			expr:    "foo(",
			wantErr: true,
		},
		{
			name:    "Missing closing parenthesis after arg",
			expr:    "foo(1",
			wantErr: true,
		},
		{
			name:    "Missing comma",
			expr:    "foo(1 2)",
			wantErr: true,
		},
		{
			name:    "Trailing comma",
			expr:    "foo(1,)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.expr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestParseCommentsLowerLevel(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name: "Comment at start",
			expr: "/* comment */ foo",
		},
		{
			name: "Comment in function call",
			expr: "foo( /* comment */ 1)",
		},
		{
			name:    "Unclosed comment error propagation",
			expr:    "foo /* unclosed",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.expr)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unclosed comment")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
