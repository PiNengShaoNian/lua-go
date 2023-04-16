package state

import "lua_go/api"

// [-0, +1, m]
// http://www.lua.org/manual/5.3/manual.html#lua_newthread
// lua-5.3.4/src/lstate.c#lua_newthread()
func (ls *luaState) NewThread() api.LuaState {
	t := &luaState{registry: ls.registry}
	t.pushLuaStack(newLuaStack(api.LUA_MINSTACK, t))
	ls.stack.push(t)
	return t
}

func (ls *luaState) Resume(from api.LuaState, nArgs int) int {
	lsFrom := from.(*luaState)
	if lsFrom.coChan == nil {
		lsFrom.coChan = make(chan int)
	}

	if ls.coChan == nil { // start coroutine
		ls.coChan = make(chan int)
		ls.coCaller = lsFrom
		go func() {
			ls.coStatus = ls.PCall(nArgs, -1, 0)
			lsFrom.coChan <- 1
		}()
	} else { // resume coroutine
		ls.coStatus = api.LUA_OK
		ls.coChan <- 1
	}

	<-lsFrom.coChan // wait coroutine to finish or yield
	return ls.coStatus
}

func (ls *luaState) Yield(nResults int) int {
	ls.coStatus = api.LUA_YIELD
	ls.coCaller.coChan <- 1
	<-ls.coChan
	return ls.GetTop()
}

func (ls *luaState) Status() int {
	return ls.coStatus
}

func (ls *luaState) GetStack() bool {
	return ls.stack.prev != nil
}

// [-0, +0, –]
// http://www.lua.org/manual/5.3/manual.html#lua_isyieldable
func (ls *luaState) IsYieldable() bool {
	if ls.isMainThread() {
		return false
	}
	return ls.coStatus != api.LUA_YIELD // todo
}
