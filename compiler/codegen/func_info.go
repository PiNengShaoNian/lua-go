package codegen

import (
	"lua_go/compiler/ast"
	"lua_go/compiler/lexer"
	"lua_go/vm"
)

var arithAndBitwiseBinops = map[int]int{
	lexer.TOKEN_OP_ADD:  vm.OP_ADD,
	lexer.TOKEN_OP_SUB:  vm.OP_SUB,
	lexer.TOKEN_OP_MUL:  vm.OP_MUL,
	lexer.TOKEN_OP_MOD:  vm.OP_MOD,
	lexer.TOKEN_OP_POW:  vm.OP_POW,
	lexer.TOKEN_OP_DIV:  vm.OP_DIV,
	lexer.TOKEN_OP_IDIV: vm.OP_IDIV,
	lexer.TOKEN_OP_BAND: vm.OP_BAND,
	lexer.TOKEN_OP_BOR:  vm.OP_BOR,
	lexer.TOKEN_OP_BXOR: vm.OP_BXOR,
	lexer.TOKEN_OP_SHL:  vm.OP_SHL,
	lexer.TOKEN_OP_SHR:  vm.OP_SHR,
}

type locVarInfo struct {
	prev     *locVarInfo
	name     string
	scopeLv  int
	slot     int
	captured bool
}

type upvalInfo struct {
	locVarSlot int
	upvalIndex int
	index      int
}

type funcInfo struct {
	constants map[interface{}]int
	usedRegs  int
	maxRegs   int
	scopeLv   int
	locVars   []*locVarInfo
	locNames  map[string]*locVarInfo
	breaks    [][]int
	parent    *funcInfo
	upvalues  map[string]upvalInfo
	insts     []uint32
	subFuncs  []*funcInfo
	numParams int
	isVararg  bool
}

func newFuncInfo(parent *funcInfo, fd *ast.FuncDefExp) *funcInfo {
	return &funcInfo{
		parent:    parent,
		subFuncs:  []*funcInfo{},
		constants: map[interface{}]int{},
		upvalues:  map[string]upvalInfo{},
		locNames:  map[string]*locVarInfo{},
		locVars:   make([]*locVarInfo, 0, 8),
		breaks:    make([][]int, 1),
		insts:     make([]uint32, 0, 8),
		isVararg:  fd.IsVararg,
		numParams: len(fd.ParList),
	}
}

func (fi *funcInfo) indexOfConstant(k interface{}) int {
	if idx, found := fi.constants[k]; found {
		return idx
	}

	idx := len(fi.constants)
	fi.constants[k] = idx

	return idx
}

func (fi *funcInfo) allocReg() int {
	fi.usedRegs++
	if fi.usedRegs >= 255 {
		panic("function or expression needs too many registers")
	}

	if fi.usedRegs > fi.maxRegs {
		fi.maxRegs = fi.usedRegs
	}

	return fi.usedRegs - 1
}

func (fi *funcInfo) freeReg() {
	if fi.usedRegs <= 0 {
		panic("usedRegs <= 0 !")
	}
	fi.usedRegs--
}

func (fi *funcInfo) allocRegs(n int) int {
	if n <= 0 {
		panic("n <= 0 !")
	}
	for i := 0; i < n; i++ {
		fi.allocReg()
	}
	return fi.usedRegs - n
}

func (fi *funcInfo) freeRegs(n int) {
	for i := 0; i < n; i++ {
		fi.freeReg()
	}
}

func (fi *funcInfo) enterScope(breakable bool) {
	fi.scopeLv++
	if breakable {
		fi.breaks = append(fi.breaks, []int{})
	} else {
		fi.breaks = append(fi.breaks, nil)
	}
}

func (fi *funcInfo) addBreakJmp(pc int) {
	for i := fi.scopeLv; i >= 0; i-- {
		if fi.breaks[i] != nil { // 循环块
			fi.breaks[i] = append(fi.breaks[i], pc)
			return
		}
	}
	panic("<break> at line ? not inside a loop!")
}

func (fi *funcInfo) addLocVar(name string) int {
	newVar := &locVarInfo{
		name:    name,
		prev:    fi.locNames[name],
		scopeLv: fi.scopeLv,
		slot:    fi.allocReg(), // 分配寄存器
	}
	fi.locVars = append(fi.locVars, newVar)
	fi.locNames[name] = newVar
	return newVar.slot
}

func (fi *funcInfo) slotOfLocVar(name string) int {
	if locVar, found := fi.locNames[name]; found {
		return locVar.slot
	}
	return -1
}

func (fi *funcInfo) exitScope() {
	pendingBreakJmps := fi.breaks[len(fi.breaks)-1]
	fi.breaks = fi.breaks[:len(fi.breaks)-1]
	a := fi.getJmpArgA()
	for _, pc := range pendingBreakJmps {
		sBx := fi.pc() - pc
		i := (sBx+vm.MAXARG_sBx)<<14 | a<<6 | vm.OP_JMP
		fi.insts[pc] = uint32(i)
	}
	fi.scopeLv--
	for _, locVar := range fi.locNames {
		if locVar.scopeLv > fi.scopeLv { // 离开作用域
			fi.removeLocVar(locVar)
		}
	}
}

func (fi *funcInfo) removeLocVar(locVar *locVarInfo) {
	fi.freeReg()
	if locVar.prev == nil {
		delete(fi.locNames, locVar.name)
	} else if locVar.prev.scopeLv == locVar.scopeLv {
		fi.removeLocVar(locVar.prev)
	} else {
		fi.locNames[locVar.name] = locVar.prev
	}
}

func (fi *funcInfo) indexOfUpval(name string) int {
	if upval, ok := fi.upvalues[name]; ok {
		return upval.index
	}

	if fi.parent != nil {
		if locVar, found := fi.parent.locNames[name]; found {
			idx := len(fi.upvalues)
			fi.upvalues[name] = upvalInfo{
				locVarSlot: locVar.slot,
				upvalIndex: -1,
				index:      idx,
			}
			locVar.captured = true
			return idx
		}
		if uvIdx := fi.parent.indexOfUpval(name); uvIdx >= 0 {
			idx := len(fi.upvalues)
			fi.upvalues[name] = upvalInfo{
				locVarSlot: -1,
				upvalIndex: uvIdx,
				index:      idx,
			}
			return idx
		}
	}

	return -1
}

func (fi *funcInfo) emitABC(opcode, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
}

func (fi *funcInfo) emitABx(opcode, a, bx int) {
	i := bx<<14 | a<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
}

func (fi *funcInfo) emitAsBx(opcode, a, b int) {
	i := (b+vm.MAXARG_sBx)<<14 | a<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
}

func (fi *funcInfo) emitAx(opcode, ax int) {
	i := ax<<6 | opcode
	fi.insts = append(fi.insts, uint32(i))
}

func (fi *funcInfo) pc() int {
	return len(fi.insts) - 1
}

func (fi *funcInfo) fixSbx(pc, sBx int) {
	i := fi.insts[pc]
	i = i << 18 >> 18                     // 清除sBx操作数
	i = i | uint32(sBx+vm.MAXARG_sBx)<<14 // 重置sBx操作数
	fi.insts[pc] = i
}

// return r[a], ... ,r[a+b-2]
func (fi *funcInfo) emitReturn(a, n int) {
	fi.emitABC(vm.OP_RETURN, a, n+1, 0)
}

// if not (r[a] <=> c) then pc++
func (fi *funcInfo) emitTest(a, c int) {
	fi.emitABC(vm.OP_TEST, a, 0, c)
}

// pc+=sBx; if (a) close all upvalues >= r[a - 1]
func (fi *funcInfo) emitJmp(a, sBx int) int {
	fi.emitAsBx(vm.OP_JMP, a, sBx)
	return len(fi.insts) - 1
}

func (fi *funcInfo) emitTForCall(a, c int) {
	fi.emitABC(vm.OP_TFORCALL, a, 0, c)
}

func (fi *funcInfo) emitTForLoop(a, sBx int) {
	fi.emitAsBx(vm.OP_TFORLOOP, a, sBx)
}

func (fi *funcInfo) emitNewTable(a, nArr, nRec int) {
	fi.emitABC(vm.OP_NEWTABLE, a, vm.Int2fb(nArr), vm.Int2fb(nRec))
}

// r[a] = op r[b]
func (fi *funcInfo) emitUnaryOp(op, a, b int) {
	switch op {
	case lexer.TOKEN_OP_NOT:
		fi.emitABC(vm.OP_NOT, a, b, 0)
	case lexer.TOKEN_OP_BNOT:
		fi.emitABC(vm.OP_BNOT, a, b, 0)
	case lexer.TOKEN_OP_LEN:
		fi.emitABC(vm.OP_LEN, a, b, 0)
	case lexer.TOKEN_OP_UNM:
		fi.emitABC(vm.OP_UNM, a, b, 0)
	}
}

// r[a] = rk[b] op rk[c]
// arith & bitwise & relational
func (fi *funcInfo) emitBinaryOp(op, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		fi.emitABC(opcode, a, b, c)
	} else {
		switch op {
		case lexer.TOKEN_OP_EQ:
			fi.emitABC(vm.OP_EQ, 1, b, c)
		case lexer.TOKEN_OP_NE:
			fi.emitABC(vm.OP_EQ, 0, b, c)
		case lexer.TOKEN_OP_LT:
			fi.emitABC(vm.OP_LT, 1, b, c)
		case lexer.TOKEN_OP_GT:
			fi.emitABC(vm.OP_LT, 1, c, b)
		case lexer.TOKEN_OP_LE:
			fi.emitABC(vm.OP_LE, 1, b, c)
		case lexer.TOKEN_OP_GE:
			fi.emitABC(vm.OP_LE, 1, c, b)
		}
		fi.emitJmp(0, 1)
		fi.emitLoadBool(a, 0, 1)
		fi.emitLoadBool(a, 1, 0)
	}
}

// r[a][rk(b)] = rk(c)
func (fi *funcInfo) emitSetTable(a, b, c int) {
	fi.emitABC(vm.OP_SETTABLE, a, b, c)
}

// r[a] := r[b][rk(c)]
func (fi *funcInfo) emitGetTable(a, b, c int) {
	fi.emitABC(vm.OP_GETTABLE, a, b, c)
}

// r[a] = (bool)b; if (c) pc++
func (fi *funcInfo) emitLoadBool(a, b, c int) {
	fi.emitABC(vm.OP_LOADBOOL, a, b, c)
}

// r[a], r[a+1], ..., r[a+b] = nil
func (fi *funcInfo) emitLoadNil(a, n int) {
	fi.emitABC(vm.OP_LOADNIL, a, n-1, 0)
}

// r[a] = r[b]
func (fi *funcInfo) emitMove(a, b int) {
	fi.emitABC(vm.OP_MOVE, a, b, 0)
}

// upval[b] = r[a]
func (fi *funcInfo) emitSetUpval(a, b int) {
	fi.emitABC(vm.OP_SETUPVAL, a, b, 0)
}

// r[a] = upval[b]
func (fi *funcInfo) emitGetUpval(a, b int) {
	fi.emitABC(vm.OP_GETUPVAL, a, b, 0)
}

// upval[a][rk(b)] = rk(c)
func (fi *funcInfo) emitSetTabUp(a, b, c int) {
	fi.emitABC(vm.OP_SETTABUP, a, b, c)
}

func (fi *funcInfo) emitForPrep(a, sBx int) int {
	fi.emitAsBx(vm.OP_FORPREP, a, sBx)
	return len(fi.insts) - 1
}

func (fi *funcInfo) emitForLoop(a, sBx int) int {
	fi.emitAsBx(vm.OP_FORLOOP, a, sBx)
	return len(fi.insts) - 1
}

// r[a] = kst[bx]
func (fi *funcInfo) emitLoadK(a int, k interface{}) {
	idx := fi.indexOfConstant(k)
	if idx < (1 << 18) {
		fi.emitABx(vm.OP_LOADK, a, idx)
	} else {
		fi.emitABx(vm.OP_LOADKX, a, 0)
		fi.emitAx(vm.OP_EXTRAARG, idx)
	}
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (fi *funcInfo) emitVararg(a, n int) {
	fi.emitABC(vm.OP_VARARG, a, n+1, 0)
}

// r[a] = emitClosure(proto[bx])
func (fi *funcInfo) emitClosure(a, bx int) {
	fi.emitABx(vm.OP_CLOSURE, a, bx)
}

// r[a][(c-1)*FPF+i] := r[a+i], 1 <= i <= b
func (fi *funcInfo) emitSetList(a, b, c int) {
	fi.emitABC(vm.OP_SETLIST, a, b, c)
}

// if (r[b] <=> c) then r[a] := r[b] else pc++
func (fi *funcInfo) emitTestSet(a, b, c int) {
	fi.emitABC(vm.OP_TESTSET, a, b, c)
}

// r[a], ..., r[a+c-2] = r[a](r[a+1], ..., r[a+b-1])
func (fi *funcInfo) emitCall(a, nArgs, nRet int) {
	fi.emitABC(vm.OP_CALL, a, nArgs+1, nRet+1)
}

// r[a+1] := r[b]; r[a] := r[b][rk(c)]
func (fi *funcInfo) emitSelf(a, b, c int) {
	fi.emitABC(vm.OP_SELF, a, b, c)
}
