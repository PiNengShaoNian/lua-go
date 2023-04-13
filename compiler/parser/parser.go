package parser

import (
	"lua_go/compiler/ast"
	"lua_go/compiler/lexer"
)

func Parse(chunk, chunkName string) *ast.Block {
	lex := lexer.NewLexer(chunk, chunkName)
	block := parseBlock(lex)
	lex.NextTokenOfKind(lexer.TOKEN_EOF)
	return block
}
