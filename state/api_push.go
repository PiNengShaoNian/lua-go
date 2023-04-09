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
	ls.stack.push(newGoClosure(f, 0))
}

func (ls *luaState) PushGlobalTable() {
	global := ls.registry.get(api.LUA_RIDX_GLOBALS)
	ls.stack.push(global)
}

func (ls *luaState) PushGoClosure(f api.GoFunction, n int) {
	closure := newGoClosure(f, n)

	for i := n; i > 0; i-- {
		val := ls.stack.pop()
		closure.upvals[n-1] = &upvalue{&val}
	}
	ls.stack.push(closure)
}
