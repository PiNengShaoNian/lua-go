package state

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
