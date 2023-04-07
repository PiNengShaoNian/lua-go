package state

import "lua_go/api"

func (ls *luaState) CreateTable(nArr, nRec int) {
	t := newLuaTable(nArr, nRec)
	ls.stack.push(t)
}

func (ls *luaState) NewTable() {
	ls.CreateTable(0, 0)
}

func (ls *luaState) GetTable(idx int) api.LuaType {
	t := ls.stack.get(idx)
	k := ls.stack.pop()
	return ls.getTable(t, k)
}

func (ls *luaState) getTable(t, k luaValue) api.LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		ls.stack.push(v)
		return typeOf(v)
	}
	panic("not a table!")
}

func (ls *luaState) GetField(idx int, k string) api.LuaType {
	t := ls.stack.get(idx)
	return ls.getTable(t, k)
}

func (ls *luaState) GetI(idx int, i int64) api.LuaType {
	t := ls.stack.get(idx)
	return ls.getTable(t, i)
}
