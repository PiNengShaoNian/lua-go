package state

import "lua_go/api"

func (ls *luaState) GetTop() int {
	return ls.stack.top
}

func (ls *luaState) AbsIndex(idx int) int {
	return ls.stack.absIndex(idx)
}

func (ls *luaState) CheckStack(n int) bool {
	ls.stack.check(n)
	return true // never fails
}

func (ls *luaState) Pop(n int) {
	ls.SetTop(-n - 1)
}

func (ls *luaState) Copy(fromIdx, toIdx int) {
	val := ls.stack.get(fromIdx)
	ls.stack.set(toIdx, val)
}

func (ls *luaState) PushValue(idx int) {
	val := ls.stack.get(idx)
	ls.stack.push(val)
}

func (ls *luaState) Replace(idx int) {
	val := ls.stack.pop()
	ls.stack.set(idx, val)
}

func (ls *luaState) Insert(idx int) {
	ls.Rotate(idx, 1)
}

func (ls *luaState) Remove(idx int) {
	ls.Rotate(idx, -1)
	ls.Pop(1)
}

func (ls *luaState) Rotate(idx, n int) {
	t := ls.stack.top - 1
	p := ls.stack.absIndex(idx - 1)
	var m int

	if n >= 0 {
		m = t - n
	} else {
		m = p - n - 1
	}
	ls.stack.reverse(p, m)
	ls.stack.reverse(m+1, t)
	ls.stack.reverse(p, t)
}

func (ls *luaState) SetTop(idx int) {
	newTop := ls.stack.absIndex(idx)
	if newTop < 0 {
		panic("stack underflow!")
	}

	n := ls.stack.top - newTop
	if n > 0 {
		for i := 0; i < n; i++ {
			ls.stack.pop()
		}
	} else if n < 0 {
		for i := 0; i > n; i-- {
			ls.stack.push(nil)
		}
	}
}

func (ls *luaState) XMove(to api.LuaState, n int) {
	vals := ls.stack.popN(n)
	to.(*luaState).stack.pushN(vals, n)
}
