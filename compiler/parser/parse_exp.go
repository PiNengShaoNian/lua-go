package parser

import (
	"lua_go/compiler/ast"
	"lua_go/compiler/lexer"
	"lua_go/number"
)

func parseExpList(lex *lexer.Lexer) []ast.Exp {
	exps := make([]ast.Exp, 0, 4)
	exps = append(exps, parseExp(lex))

	for lex.LookAhead() == lexer.TOKEN_SEP_COMMA {
		lex.NextToken()
		exps = append(exps, parseExp(lex))
	}
	return exps
}

/*
exp ::=  nil | false | true | Numeral | LiteralString | ‘...’ | functiondef |
	 prefixexp | tableconstructor | exp binop exp | unop exp
*/
/*
exp   ::= exp12
exp12 ::= exp11 {or exp11}
exp11 ::= exp10 {and exp10}
exp10 ::= exp9 {(‘<’ | ‘>’ | ‘<=’ | ‘>=’ | ‘~=’ | ‘==’) exp9}
exp9  ::= exp8 {‘|’ exp8}
exp8  ::= exp7 {‘~’ exp7}
exp7  ::= exp6 {‘&’ exp6}
exp6  ::= exp5 {(‘<<’ | ‘>>’) exp5}
exp5  ::= exp4 {‘..’ exp4}
exp4  ::= exp3 {(‘+’ | ‘-’) exp3}
exp3  ::= exp2 {(‘*’ | ‘/’ | ‘//’ | ‘%’) exp2}
exp2  ::= {(‘not’ | ‘#’ | ‘-’ | ‘~’)} exp1
exp1  ::= exp0 {‘^’ exp2}
exp0  ::= nil | false | true | Numeral | LiteralString
		| ‘...’ | functiondef | prefixexp | tableconstructor
*/
func parseExp(lexer *lexer.Lexer) ast.Exp {
	return parseExp12(lexer)
}

// x or y
func parseExp12(lex *lexer.Lexer) ast.Exp {
	exp := parseExp11(lex)
	for lex.LookAhead() == lexer.TOKEN_OP_OR {
		line, op, _ := lex.NextToken()
		lor := &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp11(lex)}
		exp = optimizeLogicalOr(lor)
	}
	return exp
}

// x and y
func parseExp11(lex *lexer.Lexer) ast.Exp {
	exp := parseExp10(lex)
	for lex.LookAhead() == lexer.TOKEN_OP_AND {
		line, op, _ := lex.NextToken()
		land := &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp10(lex)}
		exp = optimizeLogicalAnd(land)
	}
	return exp
}

// compare
func parseExp10(lex *lexer.Lexer) ast.Exp {
	exp := parseExp9(lex)
	for {
		switch lex.LookAhead() {
		case lexer.TOKEN_OP_LT, lexer.TOKEN_OP_GT, lexer.TOKEN_OP_NE,
			lexer.TOKEN_OP_LE, lexer.TOKEN_OP_GE, lexer.TOKEN_OP_EQ:
			line, op, _ := lex.NextToken()
			exp = &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp9(lex)}
		default:
			return exp
		}
	}
}

// x | y
func parseExp9(lex *lexer.Lexer) ast.Exp {
	exp := parseExp8(lex)
	for lex.LookAhead() == lexer.TOKEN_OP_BOR {
		line, op, _ := lex.NextToken()
		bor := &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp8(lex)}
		exp = optimizeBitwiseBinaryOp(bor)
	}
	return exp
}

// x ~ y
func parseExp8(lex *lexer.Lexer) ast.Exp {
	exp := parseExp7(lex)
	for lex.LookAhead() == lexer.TOKEN_OP_BXOR {
		line, op, _ := lex.NextToken()
		bxor := &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp7(lex)}
		exp = optimizeBitwiseBinaryOp(bxor)
	}
	return exp
}

// x & y
func parseExp7(lex *lexer.Lexer) ast.Exp {
	exp := parseExp6(lex)
	for lex.LookAhead() == lexer.TOKEN_OP_BAND {
		line, op, _ := lex.NextToken()
		band := &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp6(lex)}
		exp = optimizeBitwiseBinaryOp(band)
	}
	return exp
}

// shift
func parseExp6(lex *lexer.Lexer) ast.Exp {
	exp := parseExp5(lex)
	for {
		switch lex.LookAhead() {
		case lexer.TOKEN_OP_SHL, lexer.TOKEN_OP_SHR:
			line, op, _ := lex.NextToken()
			shx := &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp5(lex)}
			exp = optimizeBitwiseBinaryOp(shx)
		default:
			return exp
		}
	}
}

// a .. b
func parseExp5(lex *lexer.Lexer) ast.Exp {
	exp := parseExp4(lex)
	if lex.LookAhead() != lexer.TOKEN_OP_CONCAT {
		return exp
	}

	line := 0
	exps := []ast.Exp{exp}
	for lex.LookAhead() == lexer.TOKEN_OP_CONCAT {
		line, _, _ = lex.NextToken()
		exps = append(exps, parseExp4(lex))
	}
	return &ast.ConcatExp{Line: line, Exps: exps}
}

// x +/- y
func parseExp4(lex *lexer.Lexer) ast.Exp {
	exp := parseExp3(lex)
	for {
		switch lex.LookAhead() {
		case lexer.TOKEN_OP_ADD, lexer.TOKEN_OP_SUB:
			line, op, _ := lex.NextToken()
			arith := &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp3(lex)}
			exp = optimizeArithBinaryOp(arith)
		default:
			return exp
		}
	}
}

// *, %, /, //
func parseExp3(lex *lexer.Lexer) ast.Exp {
	exp := parseExp2(lex)
	for {
		switch lex.LookAhead() {
		case lexer.TOKEN_OP_MUL, lexer.TOKEN_OP_MOD, lexer.TOKEN_OP_DIV, lexer.TOKEN_OP_IDIV:
			line, op, _ := lex.NextToken()
			arith := &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp2(lex)}
			exp = optimizeArithBinaryOp(arith)
		default:
			return exp
		}
	}
}

// unary
func parseExp2(lex *lexer.Lexer) ast.Exp {
	switch lex.LookAhead() {
	case lexer.TOKEN_OP_UNM, lexer.TOKEN_OP_BNOT, lexer.TOKEN_OP_LEN, lexer.TOKEN_OP_NOT:
		line, op, _ := lex.NextToken()
		exp := &ast.UnopExp{Line: line, Op: op, Exp: parseExp2(lex)}
		return optimizeUnaryOp(exp)
	}
	return parseExp1(lex)
}

// x ^ y
func parseExp1(lex *lexer.Lexer) ast.Exp { // pow is right associative
	exp := parseExp0(lex)
	if lex.LookAhead() == lexer.TOKEN_OP_POW {
		line, op, _ := lex.NextToken()
		exp = &ast.BinopExp{Line: line, Op: op, Exp1: exp, Exp2: parseExp2(lex)}
	}
	return optimizePow(exp)
}

func parseExp0(lex *lexer.Lexer) ast.Exp {
	switch lex.LookAhead() {
	case lexer.TOKEN_VARARG: // ...
		line, _, _ := lex.NextToken()
		return &ast.VarargExp{Line: line}
	case lexer.TOKEN_KW_NIL: // nil
		line, _, _ := lex.NextToken()
		return &ast.NilExp{Line: line}
	case lexer.TOKEN_KW_TRUE: // true
		line, _, _ := lex.NextToken()
		return &ast.TrueExp{Line: line}
	case lexer.TOKEN_KW_FALSE: // false
		line, _, _ := lex.NextToken()
		return &ast.FalseExp{Line: line}
	case lexer.TOKEN_STRING: // LiteralString
		line, _, token := lex.NextToken()
		return &ast.StringExp{Line: line, Str: token}
	case lexer.TOKEN_NUMBER: // Numeral
		return parseNumberExp(lex)
	case lexer.TOKEN_SEP_LCURLY: // tableconstructor
		return parseTableConstructorExp(lex)
	case lexer.TOKEN_KW_FUNCTION: // functiondef
		lex.NextToken()
		return parseFuncDefExp(lex)
	default: // prefixexp
		return parsePrefixExp(lex)
	}
}

func parseNumberExp(lex *lexer.Lexer) ast.Exp {
	line, _, token := lex.NextToken()
	if i, ok := number.ParseInteger(token); ok {
		return &ast.IntegerExp{Line: line, Val: i}
	} else if f, ok := number.ParseFloat(token); ok {
		return &ast.FloatExp{Line: line, Val: f}
	} else {
		panic("not a number: " + token)
	}
}

func parseFuncDefExp(lex *lexer.Lexer) *ast.FuncDefExp {
	line := lex.Line()                                     // 关键字function已经跳过
	lex.NextTokenOfKind(lexer.TOKEN_SEP_LPAREN)            // `(`
	parList, isVararg := _parseParList(lex)                // [parlist]
	lex.NextTokenOfKind(lexer.TOKEN_SEP_RPAREN)            // `)`
	block := parseBlock(lex)                               // block
	lastLine, _ := lex.NextTokenOfKind(lexer.TOKEN_KW_END) // end
	return &ast.FuncDefExp{Line: line, LastLine: lastLine, ParList: parList, IsVararg: isVararg, Block: block}
}

func _parseParList(lex *lexer.Lexer) (names []string, isVararg bool) {
	switch lex.LookAhead() {
	case lexer.TOKEN_SEP_RPAREN:
		return nil, false
	case lexer.TOKEN_VARARG:
		lex.NextToken()
		return nil, true
	}

	_, name := lex.NextIdentifier()
	names = append(names, name)
	for lex.LookAhead() == lexer.TOKEN_SEP_COMMA {
		lex.NextToken()
		if lex.LookAhead() == lexer.TOKEN_IDENTIFIER {
			_, name := lex.NextIdentifier()
			names = append(names, name)
		} else {
			lex.NextTokenOfKind(lexer.TOKEN_VARARG)
			isVararg = true
			break
		}
	}
	return
}

func parseTableConstructorExp(lex *lexer.Lexer) *ast.TableConstructorExp {
	line := lex.Line()
	lex.NextTokenOfKind(lexer.TOKEN_SEP_LCURLY) // {
	keyExps, valExps := _parseFieldList(lex)    // [fieldlist]
	lex.NextTokenOfKind(lexer.TOKEN_SEP_RCURLY)
	lastLine := lex.Line()
	return &ast.TableConstructorExp{Line: line, LastLine: lastLine, KeyExps: keyExps, ValExps: valExps}
}

func _parseFieldList(lex *lexer.Lexer) (ks, vs []ast.Exp) {
	if lex.LookAhead() != lexer.TOKEN_SEP_RCURLY {
		k, v := _parseField(lex)
		ks = append(ks, k)
		vs = append(vs, v)
		for _isFieldSep(lex.LookAhead()) {
			lex.NextToken()
			if lex.LookAhead() != lexer.TOKEN_SEP_RCURLY {
				k, v := _parseField(lex) // field
				ks = append(ks, k)
				vs = append(vs, v)
			} else {
				break
			}
		}
	}
	return
}

func _isFieldSep(tokenKind int) bool {
	return tokenKind == lexer.TOKEN_SEP_COMMA || tokenKind == lexer.TOKEN_SEP_SEMI
}

func _parseField(lex *lexer.Lexer) (k, v ast.Exp) {
	if lex.LookAhead() == lexer.TOKEN_SEP_LBRACK {
		lex.NextToken()                             // `[`
		k = parseExp(lex)                           // exp
		lex.NextTokenOfKind(lexer.TOKEN_SEP_RBRACK) // `]`
		lex.NextTokenOfKind(lexer.TOKEN_OP_ASSIGN)  // =
		v = parseExp(lex)                           // exp
		return
	}

	exp := parseExp(lex)
	if nameExp, ok := exp.(*ast.NameExp); ok {
		if lex.LookAhead() == lexer.TOKEN_OP_ASSIGN {
			// Name `=` exp => `[` LiteralString `]` = exp
			lex.NextToken()
			k = &ast.StringExp{Line: nameExp.Line, Str: nameExp.Name}
			v = parseExp(lex)
			return
		}
	}

	return nil, exp
}

func parsePrefixExp(lex *lexer.Lexer) ast.Exp {
	var exp ast.Exp
	if lex.LookAhead() == lexer.TOKEN_IDENTIFIER {
		line, name := lex.NextIdentifier() // Name
		exp = &ast.NameExp{Line: line, Name: name}
	} else { // `(` exp `)`
		exp = parseParensExp(lex)
	}

	return _finishPrefixExp(lex, exp)
}

func _finishPrefixExp(lex *lexer.Lexer, exp ast.Exp) ast.Exp {
	for {
		switch lex.LookAhead() {
		case lexer.TOKEN_SEP_LBRACK:
			lex.NextToken()                             // `[`
			keyExp := parseExp(lex)                     // exp
			lex.NextTokenOfKind(lexer.TOKEN_SEP_RBRACK) // `]`
			exp = &ast.TableAccessExp{LastLine: lex.Line(), PrefixExp: exp, KeyExp: keyExp}
		case lexer.TOKEN_SEP_DOT:
			lex.NextToken()                    // `.`
			line, name := lex.NextIdentifier() // Name
			keyExp := &ast.StringExp{Line: line, Str: name}
			exp = &ast.TableAccessExp{LastLine: line, PrefixExp: exp, KeyExp: keyExp}
		case lexer.TOKEN_SEP_COLON, lexer.TOKEN_SEP_LPAREN, lexer.TOKEN_SEP_LCURLY, lexer.TOKEN_STRING:
			exp = _finishFuncCallExp(lex, exp) // [`:` Name] args
		default:
			return exp
		}
	}
}

func parseParensExp(lex *lexer.Lexer) ast.Exp {
	lex.NextTokenOfKind(lexer.TOKEN_SEP_LPAREN) // `(`
	exp := parseExp(lex)
	lex.NextTokenOfKind(lexer.TOKEN_SEP_RPAREN) // `)`

	switch exp.(type) {
	case *ast.VarargExp, *ast.FuncCallExp, *ast.NameExp,
		*ast.TableAccessExp:
		return &ast.ParensExp{Exp: exp}
	}

	return exp
}

func _finishFuncCallExp(lex *lexer.Lexer, prefixExp ast.Exp) *ast.FuncCallExp {
	nameExp := _parseNameExp(lex) // [`:` Name]
	line := lex.Line()            //
	args := _parseArgs(lex)       // args
	lastLine := lex.Line()        //
	return &ast.FuncCallExp{Line: line, LastLine: lastLine, PrefixExp: prefixExp, NameExp: nameExp, Args: args}
}

func _parseNameExp(lex *lexer.Lexer) *ast.StringExp {
	if lex.LookAhead() == lexer.TOKEN_SEP_COLON {
		lex.NextToken()
		line, name := lex.NextIdentifier()
		return &ast.StringExp{Line: line, Str: name}
	}
	return nil
}

func _parseArgs(lex *lexer.Lexer) (args []ast.Exp) {
	switch lex.LookAhead() {
	case lexer.TOKEN_SEP_LPAREN: // `(` [explist] `)`
		lex.NextToken()
		if lex.LookAhead() != lexer.TOKEN_SEP_RPAREN {
			args = parseExpList(lex)
		}
		lex.NextTokenOfKind(lexer.TOKEN_SEP_RPAREN)
	case lexer.TOKEN_SEP_LCURLY: // `{` [fieldlist] `}`
		args = []ast.Exp{parseTableConstructorExp(lex)}
	default: // LiteralString
		line, str := lex.NextTokenOfKind(lexer.TOKEN_STRING)
		args = []ast.Exp{&ast.StringExp{Line: line, Str: str}}
	}
	return
}
