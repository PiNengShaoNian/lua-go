package state

import (
	"lua_go/api"
	"lua_go/binchunk"
	"lua_go/vm"
)

func (ls *luaState) Load(chunk []byte, chunkName, mode string) int {
	proto := binchunk.Undump(chunk)
	c := newLuaClosure(proto)
	ls.stack.push(c)
	if len(proto.Upvalues) > 0 { // 设置_ENV
		env := ls.registry.get(api.LUA_RIDX_GLOBALS)
		c.upvals[0] = &upvalue{&env}
	}
	return 0
}

func (ls *luaState) Call(nArgs, nResults int) {
	val := ls.stack.get(-(nArgs + 1))
	if c, ok := val.(*closure); ok {
		if c.proto != nil {
			ls.callLuaClosure(nArgs, nResults, c)
		} else {
			ls.callGoClosure(nArgs, nResults, c)
		}
	} else {
		panic("not function!")
	}
}

func (ls *luaState) callLuaClosure(nArgs, nResults int, c *closure) {
	nRegs := int(c.proto.MaxStackSize)
	nParams := int(c.proto.NumParams)
	isVararg := c.proto.IsVararg == 1
	newStack := newLuaStack(nRegs+20, ls)
	newStack.closure = c

	funcAndArgs := ls.stack.popN(nArgs + 1)
	newStack.pushN(funcAndArgs[1:], nParams)
	newStack.top = nRegs
	if nArgs > nParams && isVararg {
		newStack.varargs = funcAndArgs[nParams+1:]
	}

	ls.pushLuaStack(newStack)
	ls.runLuaClosure()
	ls.popLuaStack()

	if nResults != 0 {
		results := newStack.popN(newStack.top - nRegs)
		ls.stack.check(len(results))
		ls.stack.pushN(results, nResults)
	}
}

func (ls *luaState) callGoClosure(nArgs, nResults int, c *closure) {
	newStack := newLuaStack(nArgs+20, ls)
	newStack.closure = c

	args := ls.stack.popN(nArgs)
	newStack.pushN(args, nArgs)
	ls.stack.pop()

	ls.pushLuaStack(newStack)
	r := c.goFunc(ls)
	ls.popLuaStack()

	if nResults != 0 {
		results := newStack.popN(r)
		ls.stack.check(len(results))
		ls.stack.pushN(results, nResults)
	}
}

func (ls *luaState) runLuaClosure() {
	for {
		inst := vm.Instruction(ls.Fetch())
		inst.Execute(ls)
		if inst.Opcode() == vm.OP_RETURN {
			break
		}
	}
}
