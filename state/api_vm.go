package state

func (ls *luaState) PC() int {
	return ls.stack.pc
}

func (ls *luaState) AddPC(n int) {
	ls.stack.pc += n
}

func (ls *luaState) Fetch() uint32 {
	i := ls.stack.closure.proto.Code[ls.stack.pc]
	ls.stack.pc++
	return i
}

func (ls *luaState) GetConst(idx int) {
	c := ls.stack.closure.proto.Constants[idx]
	ls.stack.push(c)
}

func (ls *luaState) GetRK(rk int) {
	if rk > 0xFF { // constant
		ls.GetConst(rk & 0xFF)
	} else { // register
		ls.PushValue(rk + 1)
	}
}

func (ls *luaState) RegisterCount() int {
	return int(ls.stack.closure.proto.MaxStackSize)
}

func (ls *luaState) LoadVararg(n int) {
	if n < 0 {
		n = len(ls.stack.varargs)
	}
	ls.stack.check(n)
	ls.stack.pushN(ls.stack.varargs, n)
}

func (ls *luaState) LoadProto(idx int) {
	proto := ls.stack.closure.proto.Protos[idx]
	closure := newLuaClosure(proto)
	ls.stack.push(closure)
}
