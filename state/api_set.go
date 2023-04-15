package state

import "lua_go/api"

func (ls *luaState) SetTable(idx int) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	k := ls.stack.pop()
	ls.setTable(t, k, v, false)
}

func (ls *luaState) setTable(t, k, v luaValue, raw bool) {
	if tbl, ok := t.(*luaTable); ok {
		if raw || tbl.get(k) != nil || !tbl.hasMetafield("__newindex") {
			tbl.put(k, v)
			return
		}
	}

	if !raw {
		if mf := getMetafield(t, "__newindex", ls); mf != nil {
			switch x := mf.(type) {
			case *luaTable:
				ls.setTable(x, k, v, false)
				return
			case *closure:
				ls.stack.push(mf)
				ls.stack.push(t)
				ls.stack.push(k)
				ls.stack.push(v)
				ls.Call(3, 0)
				return
			}
		}
	}

	panic("index error!")
}
func (ls *luaState) SetField(idx int, k string) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	ls.setTable(t, k, v, false)
}

func (ls *luaState) SetI(idx int, i int64) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	ls.setTable(t, i, v, false)
}

func (ls *luaState) SetGlobal(name string) {
	t := ls.registry.get(api.LUA_RIDX_GLOBALS)
	v := ls.stack.pop()
	ls.setTable(t, name, v, false)
}

func (ls *luaState) Register(name string, f api.GoFunction) {
	ls.PushGoFunction(f)
	ls.SetGlobal(name)
}

func (ls *luaState) SetMetatable(idx int) {
	val := ls.stack.get(idx)
	mtVal := ls.stack.pop()

	if mtVal == nil {
		setMetatable(val, nil, ls)
	} else if mt, ok := mtVal.(*luaTable); ok {
		setMetatable(val, mt, ls)
	} else {
		panic("table expected!") // todo
	}
}

func (ls *luaState) RawSet(idx int) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	k := ls.stack.pop()
	ls.setTable(t, k, v, true)
}

func (ls *luaState) RawSetI(idx int, i int64) {
	t := ls.stack.get(idx)
	v := ls.stack.pop()
	ls.setTable(t, i, v, true)
}
