package state

import "lua_go/api"

type luaState struct {
	registry *luaTable // 注册表
	stack    *luaStack
	coStatus int
	coCaller *luaState
	coChan   chan int
}

func New() *luaState {
	ls := &luaState{}

	registry := newLuaTable(8, 0)
	registry.put(api.LUA_RIDX_MAINTHREAD, ls)
	registry.put(api.LUA_RIDX_GLOBALS, newLuaTable(0, 20))

	ls.registry = registry
	ls.pushLuaStack(newLuaStack(api.LUA_MINSTACK, ls))
	return ls
}

func (ls *luaState) pushLuaStack(stack *luaStack) {
	stack.prev = ls.stack
	ls.stack = stack
}

func (ls *luaState) popLuaStack() {
	stack := ls.stack
	ls.stack = stack.prev
	stack.prev = nil
}

func (ls *luaState) isMainThread() bool {
	return ls.registry.get(api.LUA_RIDX_MAINTHREAD) == ls
}
