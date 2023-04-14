package compiler

import (
	"lua_go/binchunk"
	"lua_go/compiler/codegen"
	"lua_go/compiler/parser"
)

func Compile(chunk, chunkName string) *binchunk.Prototype {
	ast := parser.Parse(chunk, chunkName)
	return codegen.GenProto(ast)
}
