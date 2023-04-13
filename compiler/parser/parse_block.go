package parser

import (
	"lua_go/compiler/ast"
	"lua_go/compiler/lexer"
)

// block ::= {stat} [retstat]
func parseBlock(lexer *lexer.Lexer) *ast.Block {
	return &ast.Block{
		Stats:    parseStats(lexer),
		RetExps:  parseRetExps(lexer),
		LastLine: lexer.Line(),
	}
}

func parseStats(lexer *lexer.Lexer) []ast.Stat { // {stat}
	stats := make([]ast.Stat, 0, 8)
	for !_isReturnOrBlockEnd(lexer.LookAhead()) {
		stat := parseStat(lexer)
		if _, ok := stat.(*ast.EmptyStat); !ok {
			stats = append(stats, stat)
		}
	}
	return stats
}

func _isReturnOrBlockEnd(tokenKind int) bool {
	switch tokenKind {
	case lexer.TOKEN_KW_RETURN, lexer.TOKEN_EOF,
		lexer.TOKEN_KW_END, lexer.TOKEN_KW_ELSE,
		lexer.TOKEN_KW_ELSEIF, lexer.TOKEN_KW_UNTIL:
		return true
	}
	return false
}

// retstat ::= return [explist] [';']
func parseRetExps(lex *lexer.Lexer) []ast.Exp {
	if lex.LookAhead() != lexer.TOKEN_KW_RETURN {
		return nil
	}

	lex.NextToken()
	switch lex.LookAhead() {
	case lexer.TOKEN_EOF,
		lexer.TOKEN_KW_END,
		lexer.TOKEN_KW_ELSE,
		lexer.TOKEN_KW_ELSEIF,
		lexer.TOKEN_KW_UNTIL:
		return []ast.Exp{}
	case lexer.TOKEN_SEP_SEMI:
		lex.NextToken()
		return []ast.Exp{}
	default:
		exps := parseExpList(lex)
		if lex.LookAhead() == lexer.TOKEN_SEP_SEMI {
			lex.NextToken()
		}
		return exps
	}
}
