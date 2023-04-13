package parser

import (
	"encoding/json"
	"lua_go/compiler/ast"
	"testing"
)

func TestParser(t *testing.T) {
	tests := []struct {
		chunk    string
		expected *ast.Block
	}{
		{
			chunk: `print("Hello, World!")`,
			expected: &ast.Block{
				LastLine: 1,
				Stats: []ast.Stat{
					&ast.FuncCallExp{
						Line:     1,
						LastLine: 1,
						PrefixExp: &ast.NameExp{
							Line: 1,
							Name: "print",
						},
						NameExp: nil,
						Args: []ast.Exp{
							&ast.StringExp{
								Line: 1,
								Str:  "Hello, World!",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		actualBlock := Parse(tt.chunk, "")
		actualBlockJSON, err := json.Marshal(actualBlock)
		if err != nil {
			t.Fatalf(err.Error())
		}
		expectedBlockJSON, err := json.Marshal(tt.expected)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if string(expectedBlockJSON) != string(actualBlockJSON) {
			t.Fatalf("expected %q got %q", string(expectedBlockJSON), string(actualBlockJSON))
		}
	}
}
