package state

import . "lua_go/api"

type luaValue interface{}

func typeOf(val luaValue) LuaType {
	switch val.(type) {
	case nil:
		return LUA_TNIL
	case bool:
		return LUA_TBOOLEAN
	case int64, float64:
		return LUA_TNUMBER
	case string:
		return LUA_TSTRING
	default:
		panic("todo!")
	}
}
