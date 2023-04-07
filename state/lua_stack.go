package state

type luaStack struct {
	slots []luaValue
	top   int
}

func newLuaStack(size int) *luaStack {
	return &luaStack{
		slots: make([]luaValue, size),
		top:   0,
	}
}

func (ls *luaStack) check(n int) {
	free := len(ls.slots) - ls.top
	for i := free; i < n; i++ {
		ls.slots = append(ls.slots, nil)
	}
}

func (ls *luaStack) push(val luaValue) {
	if ls.top == len(ls.slots) {
		panic("stack overflow!")
	}
	ls.slots[ls.top] = val
	ls.top++
}

func (ls *luaStack) pop() luaValue {
	if ls.top < 1 {
		panic("stack underflow!")
	}
	ls.top--
	val := ls.slots[ls.top]
	ls.slots[ls.top] = nil
	return val
}

func (ls *luaStack) absIndex(idx int) int {
	if idx >= 0 {
		return idx
	}
	return idx + ls.top + 1
}

func (ls *luaStack) isValid(idx int) bool {
	absIdx := ls.absIndex(idx)
	return absIdx > 0 && absIdx <= ls.top
}

func (ls *luaStack) get(idx int) luaValue {
	absIdx := ls.absIndex(idx)
	if absIdx > 0 && absIdx <= ls.top {
		return ls.slots[absIdx-1]
	}
	return nil
}

func (ls *luaStack) set(idx int, val luaValue) {
	absIdx := ls.absIndex(idx)
	if absIdx > 0 && absIdx <= ls.top {
		ls.slots[absIdx-1] = val
		return
	}
	panic("invalid index!")
}

func (ls *luaStack) reverse(from, to int) {
	slots := ls.slots
	for from < to {
		slots[from], slots[to] = slots[to], slots[from]
		from++
		to--
	}
}
