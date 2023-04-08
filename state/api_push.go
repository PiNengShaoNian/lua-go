package state

import "lua_go/api"

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

func (ls *luaState) PushGoFunction(f api.GoFunction) {
	ls.stack.push(newGoClosure(f))
}

func (ls *luaState) PushGlobalTable() {
	global := ls.registry.get(api.LUA_RIDX_GLOBALS)
	ls.stack.push(global)
}
