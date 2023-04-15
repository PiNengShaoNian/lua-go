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
	return ls.getTable(t, k, false)
}

func (ls *luaState) getTable(t, k luaValue, raw bool) api.LuaType {
	if tbl, ok := t.(*luaTable); ok {
		v := tbl.get(k)
		if raw || v != nil || !tbl.hasMetafield("__index") {
			ls.stack.push(v)
			return typeOf(v)
		}
	}

	if !raw {
		if mf := getMetafield(t, "__index", ls); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				return ls.getTable(x, k, false)
			case *closure:
				ls.stack.push(mf)
				ls.stack.push(t)
				ls.stack.push(k)
				ls.Call(2, 1)
				v := ls.stack.get(-1)
				return typeOf(v)
			}
		}
	}

	panic("index error!")
}

func (ls *luaState) GetField(idx int, k string) api.LuaType {
	t := ls.stack.get(idx)
	return ls.getTable(t, k, false)
}

func (ls *luaState) GetI(idx int, i int64) api.LuaType {
	t := ls.stack.get(idx)
	return ls.getTable(t, i, false)
}

func (ls *luaState) GetGlobal(name string) api.LuaType {
	t := ls.registry.get(api.LUA_RIDX_GLOBALS)
	return ls.getTable(t, name, false)
}

func (ls *luaState) GetMetatable(idx int) bool {
	val := ls.stack.get(idx)

	if mt := getMetatable(val, ls); mt != nil {
		ls.stack.push(mt)
		return true
	} else {
		return false
	}
}

func (ls *luaState) RawGet(idx int) api.LuaType {
	t := ls.stack.get(idx)
	k := ls.stack.pop()
	return ls.getTable(t, k, true)
}

func (ls *luaState) RawGetI(idx int, i int64) api.LuaType {
	t := ls.stack.get(idx)
	return ls.getTable(t, i, true)
}
