package codegen

import "lua_go/compiler/ast"

func cgStat(fi *funcInfo, node ast.Stat) {
	switch stat := node.(type) {
	case *ast.FuncCallStat:
		cgFuncCallStat(fi, stat)
	case *ast.BreakStat:
		cgBreakStat(fi, stat)
	case *ast.DoStat:
		cgDoStat(fi, stat)
	case *ast.WhileStat:
		cgWhileStat(fi, stat)
	case *ast.RepeatStat:
		cgRepeatStat(fi, stat)
	case *ast.IfStat:
		cgIfStat(fi, stat)
	case *ast.ForNumStat:
		cgForNumStat(fi, stat)
	case *ast.ForInStat:
		cgForInStat(fi, stat)
	case *ast.AssignStat:
		cgAssignStat(fi, stat)
	case *ast.LocalVarDeclStat:
		cgLocalVarDeclStat(fi, stat)
	case *ast.LocalFuncDefStat:
		cgLocalFuncDefStat(fi, stat)
	case *ast.LabelStat, *ast.GotoStat:
		panic("label and goto statements are not supported!")
	}
}

func cgLocalFuncDefStat(fi *funcInfo, node *ast.LocalFuncDefStat) {
	r := fi.addLocVar(node.Name)
	cgFuncDefExp(fi, node.Exp, r)
}

func cgFuncCallStat(fi *funcInfo, node *ast.FuncCallStat) {
	r := fi.allocReg()
	cgFuncCallExp(fi, node, r, 0)
	fi.freeReg()
}

func cgBreakStat(fi *funcInfo, node *ast.BreakStat) {
	pc := fi.emitJmp(0, 0)
	fi.addBreakJmp(pc)
}

func cgDoStat(fi *funcInfo, node *ast.DoStat) {
	fi.enterScope(false) // 非循环块
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	fi.exitScope()
}

func (fi *funcInfo) closeOpenUpvals() {
	a := fi.getJmpArgA()
	if a > 0 {
		fi.emitJmp(a, 0)
	}
}

func (fi *funcInfo) getJmpArgA() int {
	hasCapturedLocVars := false
	minSlotOfLocVars := fi.maxRegs
	for _, locVar := range fi.locNames {
		if locVar.scopeLv == fi.scopeLv {
			for v := locVar; v != nil && v.scopeLv == fi.scopeLv; v = v.prev {
				if v.captured {
					hasCapturedLocVars = true
				}
				if v.slot < minSlotOfLocVars && v.name[0] != '(' {
					minSlotOfLocVars = v.slot
				}
			}
		}
	}
	if hasCapturedLocVars {
		return minSlotOfLocVars + 1
	} else {
		return 0
	}
}

/*
           ______________
          /  false? jmp  |
         /               |
while exp do block end <-'
      ^           \
      |___________/
           jmp
*/
func cgWhileStat(fi *funcInfo, node *ast.WhileStat) {
	pcBeforeExp := fi.pc()

	oldRegs := fi.usedRegs
	a, _ := expToOpArg(fi, node.Exp, ARG_REG)
	fi.usedRegs = oldRegs

	fi.emitTest(a, 0)
	pcJmpToEnd := fi.emitJmp(0, 0)

	fi.enterScope(true)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	fi.emitJmp(0, pcBeforeExp-fi.pc()-1)
	fi.exitScope()

	fi.fixSbx(pcJmpToEnd, fi.pc()-pcJmpToEnd)
}

/*
        ______________
       |  false? jmp  |
       V              /
repeat block until exp
*/
func cgRepeatStat(fi *funcInfo, node *ast.RepeatStat) {
	fi.enterScope(true)

	pcBeforeBlock := fi.pc()
	cgBlock(fi, node.Block)

	oldRegs := fi.usedRegs
	a, _ := expToOpArg(fi, node.Exp, ARG_REG)
	fi.usedRegs = oldRegs

	fi.emitTest(a, 0)
	fi.emitJmp(fi.getJmpArgA(), pcBeforeBlock-fi.pc()-1)
	fi.closeOpenUpvals()

	fi.exitScope()
}

/*
         _________________       _________________       _____________
        / false? jmp      |     / false? jmp      |     / false? jmp  |
       /                  V    /                  V    /              V
if exp1 then block1 elseif exp2 then block2 elseif true then block3 end <-.
                   \                       \                       \      |
                    \_______________________\_______________________\_____|
                    jmp                     jmp                     jmp
*/
func cgIfStat(fi *funcInfo, node *ast.IfStat) {
	pcJmpToEnds := make([]int, len(node.Exps))
	pcJmpToNextExp := -1

	for i, exp := range node.Exps {
		if pcJmpToNextExp >= 0 {
			fi.fixSbx(pcJmpToNextExp, fi.pc()-pcJmpToNextExp)
		}

		oldRegs := fi.usedRegs
		a, _ := expToOpArg(fi, exp, ARG_REG)
		fi.usedRegs = oldRegs

		fi.emitTest(a, 0)
		pcJmpToNextExp = fi.emitJmp(0, 0)

		block := node.Blocks[i]
		fi.enterScope(false)
		cgBlock(fi, block)
		fi.closeOpenUpvals()
		fi.exitScope()
		if i < len(node.Exps)-1 {
			pcJmpToEnds[i] = fi.emitJmp(0, 0)
		} else {
			pcJmpToEnds[i] = pcJmpToNextExp
		}
	}

	for _, pc := range pcJmpToEnds {
		fi.fixSbx(pc, fi.pc()-pc)
	}
}

func cgForNumStat(fi *funcInfo, node *ast.ForNumStat) {
	fi.enterScope(true)

	cgLocalVarDeclStat(fi, &ast.LocalVarDeclStat{
		NameList: []string{"(for index)", "(for limit)", "(for step)"},
		ExpList:  []ast.Exp{node.InitExp, node.LimitExp, node.StepExp},
	})
	fi.addLocVar(node.VarName)

	a := fi.usedRegs - 4
	pcForPrep := fi.emitForPrep(a, 0)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	pcForLoop := fi.emitForLoop(a, 0)

	fi.fixSbx(pcForPrep, pcForLoop-pcForPrep-1)
	fi.fixSbx(pcForLoop, pcForPrep-pcForLoop)

	fi.exitScope()
}

func cgForInStat(fi *funcInfo, node *ast.ForInStat) {
	fi.enterScope(true)

	cgLocalVarDeclStat(fi, &ast.LocalVarDeclStat{
		NameList: []string{"(for generator)", "(for state)", "(for control)"},
		ExpList:  node.ExpList,
	})

	for _, name := range node.NameList {
		fi.addLocVar(name)
	}

	pcJmpToTFC := fi.emitJmp(0, 0)
	cgBlock(fi, node.Block)
	fi.closeOpenUpvals()
	fi.fixSbx(pcJmpToTFC, fi.pc()-pcJmpToTFC)

	rGenerator := fi.slotOfLocVar("(for generator)")
	fi.emitTForCall(rGenerator, len(node.NameList))
	fi.emitTForLoop(rGenerator+2, pcJmpToTFC-fi.pc()-1)

	fi.exitScope()
}

func cgLocalVarDeclStat(fi *funcInfo, node *ast.LocalVarDeclStat) {
	nExps := len(node.ExpList)
	nNames := len(node.NameList)

	oldRegs := fi.usedRegs
	if nExps == nNames {
		for _, exp := range node.ExpList {
			a := fi.allocReg()
			cgExp(fi, exp, a, 1)
		}
	} else if nExps > nNames {
		for i, exp := range node.ExpList {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else { // nNames > nExps
		multRet := false
		for i, exp := range node.ExpList {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nNames - nExps + 1
				cgExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
		if !multRet {
			n := nNames - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(a, n)
		}
	}
	fi.usedRegs = oldRegs
	for _, name := range node.NameList {
		fi.addLocVar(name)
	}
}

func cgAssignStat(fi *funcInfo, node *ast.AssignStat) {
	nExps := len(node.ExpList)
	nVars := len(node.VarList)
	oldRegs := fi.usedRegs
	tRegs := make([]int, nVars)
	kRegs := make([]int, nVars)
	vRegs := make([]int, nVars)

	for i, exp := range node.VarList {
		if taExp, ok := exp.(*ast.TableAccessExp); ok {
			tRegs[i] = fi.allocReg()
			cgExp(fi, taExp.PrefixExp, tRegs[i], 1)
			kRegs[i] = fi.allocReg()
			cgExp(fi, taExp.KeyExp, kRegs[i], 1)
		}
	}

	for i := 0; i < nVars; i++ {
		vRegs[i] = fi.usedRegs + i
	}

	if nExps >= nVars {
		for i, exp := range node.ExpList {
			a := fi.allocReg()
			if i >= nVars && i == nExps-1 && isVarargOrFuncCall(exp) {
				cgExp(fi, exp, a, 0)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}
	} else { // nVars > nExps
		multRet := false

		for i, exp := range node.ExpList {
			a := fi.allocReg()
			if i == nExps-1 && isVarargOrFuncCall(exp) {
				multRet = true
				n := nVars - nExps + 1
				cgExp(fi, exp, a, n)
				fi.allocRegs(n - 1)
			} else {
				cgExp(fi, exp, a, 1)
			}
		}

		if !multRet {
			n := nVars - nExps
			a := fi.allocRegs(n)
			fi.emitLoadNil(a, n)
		}
	}

	for i, exp := range node.VarList {
		if nameExp, ok := exp.(*ast.NameExp); ok {
			varName := nameExp.Name
			if a := fi.slotOfLocVar(varName); a >= 0 {
				fi.emitMove(a, vRegs[i])
			} else if b := fi.indexOfUpval(varName); b >= 0 {
				fi.emitSetUpval(vRegs[i], b)
			} else { // global var
				a := fi.indexOfUpval("_ENV")
				b := 0x100 + fi.indexOfConstant(varName)
				fi.emitSetTabUp(a, b, vRegs[i])
			}
		} else {
			fi.emitSetTable(tRegs[i], kRegs[i], vRegs[i])
		}
	}
	fi.usedRegs = oldRegs
}
