package state

import "lua_go/api"

func (ls *luaState) SetTable(idx int) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	k := ls.stack.pop()
	ls.setTable(t, k, v)
}

func (ls *luaState) setTable(t, k, v luaValue) {
	if tbl, ok := t.(*luaTable); ok {
		tbl.put(k, v)
		return
	}
	panic("not a table!")
}

func (ls *luaState) SetField(idx int, k string) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	ls.setTable(t, k, v)
}

func (ls *luaState) SetI(idx int, i int64) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	ls.setTable(t, i, v)
}

func (ls *luaState) SetGlobal(name string) {
	t := ls.registry.get(api.LUA_RIDX_GLOBALS)
	v := ls.stack.pop()
	ls.setTable(t, name, v)
}

func (ls *luaState) Register(name string, f api.GoFunction) {
	ls.PushGoFunction(f)
	ls.SetGlobal(name)
}
