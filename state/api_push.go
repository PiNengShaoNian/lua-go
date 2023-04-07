package state

func (ls *luaState) PushNil() {
	ls.stack.push(nil)
}

func (ls *luaState) PushBoolean(b bool) {
	ls.stack.push(b)
}

func (ls *luaState) PushInteger(n int64) {
	ls.stack.push(n)
}

func (ls *luaState) PushNumber(n float64) {
	ls.stack.push(n)
}

func (ls *luaState) PushString(s string) {
	ls.stack.push(s)
}
