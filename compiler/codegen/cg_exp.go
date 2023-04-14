package codegen

import (
	"lua_go/compiler/ast"
	"lua_go/compiler/lexer"
	"lua_go/vm"
)

func cgExp(fi *funcInfo, node ast.Exp, a, n int) {
	switch exp := node.(type) {
	case *ast.NilExp:
		fi.emitLoadNil(a, n)
	case *ast.FalseExp:
		fi.emitLoadBool(a, 0, 0)
	case *ast.TrueExp:
		fi.emitLoadBool(a, 1, 0)
	case *ast.IntegerExp:
		fi.emitLoadK(a, exp.Val)
	case *ast.FloatExp:
		fi.emitLoadK(a, exp.Val)
	case *ast.StringExp:
		fi.emitLoadK(a, exp.Str)
	case *ast.ParensExp:
		cgExp(fi, exp.Exp, a, 1)
	case *ast.VarargExp:
		cgVarargExp(fi, exp, a, n)
	case *ast.FuncDefExp:
		cgFuncDefExp(fi, exp, a)
	case *ast.TableConstructorExp:
		cgTableConstructorExp(fi, exp, a)
	case *ast.UnopExp:
		cgUnopExp(fi, exp, a)
	case *ast.BinopExp:
		cgBinopExp(fi, exp, a)
	case *ast.ConcatExp:
		cgConcatExp(fi, exp, a)
	case *ast.NameExp:
		cgNameExp(fi, exp, a)
	case *ast.TableAccessExp:
		cgTableAccessExp(fi, exp, a)
	case *ast.FuncCallExp:
		cgFuncCallExp(fi, exp, a, n)
	}
}

func cgVarargExp(fi *funcInfo, node *ast.VarargExp, a, n int) {
	if !fi.isVararg {
		panic("cannot use '...' outside a vararg function")
	}
	fi.emitVararg(a, n)
}

func cgFuncDefExp(fi *funcInfo, node *ast.FuncDefExp, a int) {
	subFI := newFuncInfo(fi, node)
	fi.subFuncs = append(fi.subFuncs, subFI)

	for _, param := range node.ParList {
		subFI.addLocVar(param)
	}
	cgBlock(subFI, node.Block)
	subFI.exitScope()
	subFI.emitReturn(0, 0)

	bx := len(fi.subFuncs) - 1
	fi.emitClosure(a, bx)
}

func cgTableConstructorExp(fi *funcInfo, node *ast.TableConstructorExp, a int) {
	nArr := 0
	for _, keyExp := range node.KeyExps {
		if keyExp == nil {
			nArr++
		}
	}
	nExps := len(node.KeyExps)
	multRet := nExps > 0 &&
		isVarargOrFuncCall(node.ValExps[nExps-1])

	fi.emitNewTable(a, nArr, nExps-nArr)

	arrIdx := 0
	for i, keyExp := range node.KeyExps {
		valExp := node.ValExps[i]

		if keyExp == nil {
			arrIdx++
			tmp := fi.allocReg()
			if i == nExps-1 && multRet {
				cgExp(fi, valExp, tmp, -1)
			} else {
				cgExp(fi, valExp, tmp, 1)
			}

			if arrIdx%50 == 0 || arrIdx == nArr { // LFIELDS_PER_FLUSH
				n := arrIdx % 50
				if n == 0 {
					n = 50
				}
				fi.freeRegs(n)
				c := (arrIdx-1)/50 + 1 // todo: c > 0xFF
				if i == nExps-1 && multRet {
					fi.emitSetList(a, 0, c)
				} else {
					fi.emitSetList(a, n, c)
				}
			}

			continue
		}

		b := fi.allocReg()
		cgExp(fi, keyExp, b, 1)
		c := fi.allocReg()
		cgExp(fi, valExp, c, 1)
		fi.freeRegs(2)

		fi.emitSetTable(a, b, c)
	}
}

func cgUnopExp(fi *funcInfo, node *ast.UnopExp, a int) {
	b := fi.allocReg()
	cgExp(fi, node.Exp, b, 1)
	fi.emitUnaryOp(node.Op, a, b)
	fi.freeReg()
}

func cgConcatExp(fi *funcInfo, node *ast.ConcatExp, a int) {
	for _, subExp := range node.Exps {
		a := fi.allocReg()
		cgExp(fi, subExp, a, 1)
	}

	c := fi.usedRegs - 1
	b := c - len(node.Exps) + 1
	fi.freeRegs(c - b + 1)
	fi.emitABC(vm.OP_CONCAT, a, b, c)
}

func cgBinopExp(fi *funcInfo, node *ast.BinopExp, a int) {
	switch node.Op {
	case lexer.TOKEN_OP_AND, lexer.TOKEN_OP_OR:
		b := fi.allocReg()
		cgExp(fi, node.Exp1, b, 1)
		fi.freeReg()
		if node.Op == lexer.TOKEN_OP_AND {
			fi.emitTestSet(a, b, 0)
		} else {
			fi.emitTestSet(a, b, 1)
		}
		pcOfJmp := fi.emitJmp(0, 0)

		b = fi.allocReg()
		cgExp(fi, node.Exp2, b, 1)
		fi.freeReg()

		fi.emitMove(a, b)
		fi.fixSbx(pcOfJmp, fi.pc()-pcOfJmp)
	default:
		b := fi.allocReg()
		cgExp(fi, node.Exp1, b, 1)
		c := fi.allocReg()
		cgExp(fi, node.Exp2, c, 1)
		fi.emitBinaryOp(node.Op, a, b, c)
		fi.freeRegs(2)
	}
}

func cgNameExp(fi *funcInfo, node *ast.NameExp, a int) {
	if r := fi.slotOfLocVar(node.Name); r >= 0 {
		fi.emitMove(a, r)
	} else if idx := fi.indexOfUpval(node.Name); idx >= 0 {
		fi.emitGetUpval(a, idx)
	} else { // x => _ENV["x"]
		taExp := &ast.TableAccessExp{
			PrefixExp: &ast.NameExp{Line: 0, Name: "_ENV"},
			KeyExp:    &ast.StringExp{Line: 0, Str: node.Name},
		}

		cgTableAccessExp(fi, taExp, a)
	}
}

func cgTableAccessExp(fi *funcInfo, node *ast.TableAccessExp, a int) {
	b := fi.allocReg()
	cgExp(fi, node.PrefixExp, b, 1)
	c := fi.allocReg()
	cgExp(fi, node.KeyExp, c, 1)
	fi.emitGetTable(a, b, c)
	fi.freeRegs(2)
}

func cgFuncCallExp(fi *funcInfo, node *ast.FuncCallExp, a, n int) {
	nArgs := prepFuncCall(fi, node, a)
	fi.emitCall(a, nArgs, n)
}

func prepFuncCall(fi *funcInfo, node *ast.FuncCallExp, a int) int {
	nArgs := len(node.Args)
	lastArgIsVarargOrFuncCall := false

	cgExp(fi, node.PrefixExp, a, 1)
	if node.NameExp != nil {
		c := 0x100 + fi.indexOfConstant(node.NameExp.Str)
		fi.emitSelf(a, a, c)
	}
	for i, arg := range node.Args {
		tmp := fi.allocReg()
		if i == nArgs-1 && isVarargOrFuncCall(arg) {
			lastArgIsVarargOrFuncCall = true
			cgExp(fi, arg, tmp, -1)
		} else {
			cgExp(fi, arg, tmp, 1)
		}
	}
	fi.freeRegs(nArgs)

	if node.NameExp != nil {
		nArgs++
	}
	if lastArgIsVarargOrFuncCall {
		nArgs = -1
	}

	return nArgs
}
