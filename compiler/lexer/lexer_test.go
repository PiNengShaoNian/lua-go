package lexer

import (
	"fmt"
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		chunk    string
		expected []string
	}{
		{
			chunk: `print("Hello, World!")`,
			expected: []string{
				"[identifier] print",
				"[separator] (",
				"[string] Hello, World!",
				"[separator] )",
				"[other] EOF",
			},
		},
	}

	for _, tt := range tests {
		lexer := NewLexer(tt.chunk, "")
		for _, token := range tt.expected {
			_, kind, actualToken := lexer.NextToken()
			actual := fmt.Sprintf("[%s] %s", kindToCategory(kind), actualToken)
			if token != actual {
				t.Fatalf("expected %q got %q", token, actual)
			}
		}
	}
}

func kindToCategory(kind int) string {
	switch {
	case kind < TOKEN_SEP_SEMI:
		return "other"
	case kind <= TOKEN_SEP_RCURLY:
		return "separator"
	case kind <= TOKEN_OP_NOT:
		return "operator"
	case kind <= TOKEN_KW_WHILE:
		return "keyword"
	case kind == TOKEN_IDENTIFIER:
		return "identifier"
	case kind == TOKEN_NUMBER:
		return "number"
	case kind == TOKEN_STRING:
		return "string"
	default:
		return "other"
	}
}
