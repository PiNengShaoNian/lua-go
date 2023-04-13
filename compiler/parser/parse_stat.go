package parser

import (
	"lua_go/compiler/ast"
	"lua_go/compiler/lexer"
)

/*
stat ::=  ‘;’

	| break
	| ‘::’ Name ‘::’
	| goto Name
	| do block end
	| while exp do block end
	| repeat block until exp
	| if exp then block {elseif exp then block} [else block] end
	| for Name ‘=’ exp ‘,’ exp [‘,’ exp] do block end
	| for namelist in explist do block end
	| function funcname funcbody
	| local function Name funcbody
	| local namelist [‘=’ explist]
	| varlist ‘=’ explist
	| functioncall
*/
func parseStat(lex *lexer.Lexer) ast.Stat {
	switch lex.LookAhead() {
	case lexer.TOKEN_SEP_SEMI:
		return parseEmptyStat(lex)
	case lexer.TOKEN_KW_BREAK:
		return parseBreakStat(lex)
	case lexer.TOKEN_SEP_LABEL:
		return parseLabelStat(lex)
	case lexer.TOKEN_KW_GOTO:
		return parseGotoStat(lex)
	case lexer.TOKEN_KW_DO:
		return parseDoStat(lex)
	case lexer.TOKEN_KW_WHILE:
		return parseWhileStat(lex)
	case lexer.TOKEN_KW_REPEAT:
		return parseRepeatStat(lex)
	case lexer.TOKEN_KW_IF:
		return parseIfStat(lex)
	case lexer.TOKEN_KW_FOR:
		return parseForStat(lex)
	case lexer.TOKEN_KW_FUNCTION:
		return parseFuncDefStat(lex)
	case lexer.TOKEN_KW_LOCAL:
		return parseLocalAssignOrFuncDefStat(lex)
	default:
		return parseAssignOrFuncCallStat(lex)
	}
}

func parseEmptyStat(lex *lexer.Lexer) *ast.EmptyStat {
	lex.NextTokenOfKind(lexer.TOKEN_SEP_SEMI) // `;`
	return &ast.EmptyStat{}
}

func parseBreakStat(lex *lexer.Lexer) *ast.BreakStat {
	lex.NextTokenOfKind(lexer.TOKEN_KW_BREAK) // break
	return &ast.BreakStat{Line: lex.Line()}
}

func parseLabelStat(lex *lexer.Lexer) *ast.LabelStat {
	lex.NextTokenOfKind(lexer.TOKEN_SEP_LABEL) // `::`
	_, name := lex.NextIdentifier()            // Name
	lex.NextTokenOfKind(lexer.TOKEN_SEP_LABEL) // `::`
	return &ast.LabelStat{Name: name}
}

func parseGotoStat(lex *lexer.Lexer) *ast.GotoStat {
	lex.NextTokenOfKind(lexer.TOKEN_KW_GOTO) // goto
	_, name := lex.NextIdentifier()          // Name
	return &ast.GotoStat{Name: name}
}

func parseDoStat(lex *lexer.Lexer) *ast.DoStat {
	lex.NextTokenOfKind(lexer.TOKEN_KW_DO)  // do
	block := parseBlock(lex)                // block
	lex.NextTokenOfKind(lexer.TOKEN_KW_END) // end
	return &ast.DoStat{Block: block}
}

func parseWhileStat(lex *lexer.Lexer) *ast.WhileStat {
	lex.NextTokenOfKind(lexer.TOKEN_KW_WHILE) // while
	exp := parseExp(lex)                      // exp
	lex.NextTokenOfKind(lexer.TOKEN_KW_DO)    // do
	block := parseBlock(lex)                  // block
	lex.NextTokenOfKind(lexer.TOKEN_KW_END)   // end
	return &ast.WhileStat{Exp: exp, Block: block}
}

func parseRepeatStat(lex *lexer.Lexer) *ast.RepeatStat {
	lex.NextTokenOfKind(lexer.TOKEN_KW_REPEAT) // repeat
	block := parseBlock(lex)                   // block
	lex.NextTokenOfKind(lexer.TOKEN_KW_UNTIL)  // until
	exp := parseExp(lex)
	return &ast.RepeatStat{Block: block, Exp: exp}
}

func parseIfStat(lex *lexer.Lexer) *ast.IfStat {
	exps := make([]ast.Exp, 0, 4)
	blocks := make([]*ast.Block, 0, 4)

	lex.NextTokenOfKind(lexer.TOKEN_KW_IF)   // if
	exps = append(exps, parseExp(lex))       // exp
	lex.NextTokenOfKind(lexer.TOKEN_KW_THEN) // then
	blocks = append(blocks, parseBlock(lex)) // block

	for lex.LookAhead() == lexer.TOKEN_KW_ELSEIF { // {
		lex.NextToken()                          // elseif
		exps = append(exps, parseExp(lex))       // exp
		lex.NextTokenOfKind(lexer.TOKEN_KW_THEN) // then
		blocks = append(blocks, parseBlock(lex)) // block
	}

	// else block => elseif true then block
	if lex.LookAhead() == lexer.TOKEN_KW_ELSE {
		lex.NextToken() // else
		exps = append(exps, &ast.TrueExp{Line: lex.Line()})
		blocks = append(blocks, parseBlock(lex)) // block
	}

	lex.NextTokenOfKind(lexer.TOKEN_KW_END) // end
	return &ast.IfStat{Exps: exps, Blocks: blocks}
}

func parseForStat(lex *lexer.Lexer) ast.Stat {
	lineOfFor, _ := lex.NextTokenOfKind(lexer.TOKEN_KW_FOR)
	_, name := lex.NextIdentifier()
	if lex.LookAhead() == lexer.TOKEN_OP_ASSIGN {
		return _finishForNumStat(lex, lineOfFor, name)
	} else {
		return _finishForInStat(lex, name)
	}
}

func _finishForNumStat(lex *lexer.Lexer, lineOfFor int, varName string) *ast.ForNumStat {
	lex.NextTokenOfKind(lexer.TOKEN_OP_ASSIGN) // for name =
	initExp := parseExp(lex)                   // exp
	lex.NextTokenOfKind(lexer.TOKEN_SEP_COMMA) // `,`
	limitExp := parseExp(lex)                  // exp

	var stepExp ast.Exp
	if lex.LookAhead() == lexer.TOKEN_SEP_COMMA { // [
		lex.NextToken()         // `,`
		stepExp = parseExp(lex) // exp
	} else {
		stepExp = &ast.IntegerExp{Line: lex.Line(), Val: 1}
	}

	lineOfDo, _ := lex.NextTokenOfKind(lexer.TOKEN_KW_DO) // do
	block := parseBlock(lex)
	lex.NextTokenOfKind(lexer.TOKEN_KW_END)

	return &ast.ForNumStat{
		LineOfFor: lineOfFor,
		LineOfDo:  lineOfDo,
		VarName:   varName,
		InitExp:   initExp,
		LimitExp:  limitExp,
		StepExp:   stepExp,
		Block:     block,
	}
}

func _finishForInStat(lex *lexer.Lexer, name0 string) *ast.ForInStat {
	nameList := _finishNameList(lex, name0)               // for namelist
	lex.NextTokenOfKind(lexer.TOKEN_KW_IN)                // in
	expList := parseExpList(lex)                          // explist
	lineOfDo, _ := lex.NextTokenOfKind(lexer.TOKEN_KW_DO) // do
	block := parseBlock(lex)                              // block
	lex.NextTokenOfKind(lexer.TOKEN_KW_END)               // end

	return &ast.ForInStat{LineOfDo: lineOfDo, NameList: nameList, ExpList: expList, Block: block}
}

func _finishNameList(lex *lexer.Lexer, name0 string) []string {
	names := []string{name0}                       // Name
	for lex.LookAhead() == lexer.TOKEN_SEP_COMMA { // {
		lex.NextToken()                 // `,`
		_, name := lex.NextIdentifier() // Name
		names = append(names, name)
	}
	return names
}

func parseLocalAssignOrFuncDefStat(lex *lexer.Lexer) ast.Stat {
	lex.NextTokenOfKind(lexer.TOKEN_KW_LOCAL)
	if lex.LookAhead() == lexer.TOKEN_KW_FUNCTION {
		return _finishLocalFuncDefStat(lex)
	} else {
		return _finishLocalVarDeclStat(lex)
	}
}

func _finishLocalFuncDefStat(lex *lexer.Lexer) *ast.LocalFuncDefStat {
	lex.NextTokenOfKind(lexer.TOKEN_KW_FUNCTION) // local function
	_, name := lex.NextIdentifier()
	fdExp := parseFuncDefExp(lex) // funcbody
	return &ast.LocalFuncDefStat{Name: name, Exp: fdExp}
}

func _finishLocalVarDeclStat(lex *lexer.Lexer) *ast.LocalVarDeclStat {
	_, name0 := lex.NextIdentifier()        // local name
	nameList := _finishNameList(lex, name0) // {`,` Name}
	var expList []ast.Exp = nil
	if lex.LookAhead() == lexer.TOKEN_OP_ASSIGN { // [
		lex.NextToken()             // `=`
		expList = parseExpList(lex) // explist
	}
	lastLine := lex.Line()
	return &ast.LocalVarDeclStat{LastLine: lastLine, NameList: nameList, ExpList: expList}
}

func parseAssignOrFuncCallStat(lex *lexer.Lexer) ast.Stat {
	prefixExp := parsePrefixExp(lex)
	if fc, ok := prefixExp.(*ast.FuncCallExp); ok {
		return fc
	} else {
		return parseAssignStat(lex, prefixExp)
	}
}

func parseAssignStat(lex *lexer.Lexer, var0 ast.Exp) *ast.AssignStat {
	varList := _finishVarList(lex, var0)       // varlist
	lex.NextTokenOfKind(lexer.TOKEN_OP_ASSIGN) // `=`
	expList := parseExpList(lex)               // explist
	lastLine := lex.Line()
	return &ast.AssignStat{LastLine: lastLine, VarList: varList, ExpList: expList}
}

func _finishVarList(lex *lexer.Lexer, var0 ast.Exp) []ast.Exp {
	vars := []ast.Exp{_checkVar(lex, var0)}        // var
	for lex.LookAhead() == lexer.TOKEN_SEP_COMMA { // {
		lex.NextToken()            // `,`
		exp := parsePrefixExp(lex) // var
		vars = append(vars, _checkVar(lex, exp))
	}
	return vars
}

func _checkVar(lex *lexer.Lexer, exp ast.Exp) ast.Exp {
	switch exp.(type) {
	case *ast.NameExp, *ast.TableAccessExp:
		return exp
	}
	lex.NextTokenOfKind(-1) // trigger error
	panic("unreachable!")
}

func parseFuncDefStat(lex *lexer.Lexer) *ast.AssignStat {
	lex.NextTokenOfKind(lexer.TOKEN_KW_FUNCTION) // function
	fnExp, hasColon := _parseFuncName(lex)       // funcname
	fdExp := parseFuncDefExp(lex)                // funcbody
	if hasColon {
		fdExp.ParList = append(fdExp.ParList, "")
		copy(fdExp.ParList[1:], fdExp.ParList)
		fdExp.ParList[0] = "self"
	}

	return &ast.AssignStat{
		LastLine: fdExp.Line,
		VarList:  []ast.Exp{fnExp},
		ExpList:  []ast.Exp{fdExp},
	}
}

func _parseFuncName(lex *lexer.Lexer) (exp ast.Exp, hasColon bool) {
	line, name := lex.NextIdentifier()
	exp = &ast.NameExp{Line: line, Name: name}

	for lex.LookAhead() == lexer.TOKEN_SEP_DOT {
		lex.NextToken()
		line, name := lex.NextIdentifier()
		idx := &ast.StringExp{Line: line, Str: name}
		exp = &ast.TableAccessExp{LastLine: line, PrefixExp: exp, KeyExp: idx}
	}
	if lex.LookAhead() == lexer.TOKEN_SEP_COLON {
		lex.NextToken()
		line, name := lex.NextIdentifier()
		idx := &ast.StringExp{Line: line, Str: name}
		exp = &ast.TableAccessExp{LastLine: line, PrefixExp: exp, KeyExp: idx}
		hasColon = true
	}

	return
}
