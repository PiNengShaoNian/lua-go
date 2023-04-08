package state

type luaState struct {
	stack *luaStack
}

func New() *luaState {
	return &luaState{
		stack: newLuaStack(20),
	}
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
