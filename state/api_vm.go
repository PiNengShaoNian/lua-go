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
	stack := ls.stack
	subProto := stack.closure.proto.Protos[idx]
	closure := newLuaClosure(subProto)
	stack.push(closure)

	for i, uvInfo := range subProto.Upvalues {
		uvIdx := int(uvInfo.Idx)
		if uvInfo.Instack == 1 {
			if stack.openuvs == nil {
				stack.openuvs = map[int]*upvalue{}
			}
			if openuv, found := stack.openuvs[uvIdx]; found {
				closure.upvals[i] = openuv
			} else {
				closure.upvals[i] = &upvalue{&stack.slots[uvIdx]}
				stack.openuvs[uvIdx] = closure.upvals[i]
			}
		} else {
			closure.upvals[i] = stack.closure.upvals[uvIdx]
		}
	}
}

func (ls *luaState) CloseUpvalues(a int) {
	for i, openuv := range ls.stack.openuvs {
		if i >= a-1 {
			val := *openuv.val
			openuv.val = &val
			delete(ls.stack.openuvs, i)
		}
	}
}
