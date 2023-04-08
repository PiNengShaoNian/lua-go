package state

import (
	"fmt"
	"lua_go/api"
)

func (ls *luaState) TypeName(tp api.LuaType) string {
	switch tp {
	case api.LUA_TNONE:
		return "no value"
	case api.LUA_TNIL:
		return "nil"
	case api.LUA_TBOOLEAN:
		return "boolean"
	case api.LUA_TNUMBER:
		return "number"
	case api.LUA_TSTRING:
		return "string"
	case api.LUA_TTABLE:
		return "table"
	case api.LUA_TFUNCTION:
		return "function"
	case api.LUA_TTHREAD:
		return "thread"
	default:
		return "userdata"
	}
}

func (ls *luaState) Type(idx int) api.LuaType {
	if ls.stack.isValid(idx) {
		val := ls.stack.get(idx)
		return typeOf(val)
	}
	return api.LUA_TNONE
}

func (ls *luaState) IsNone(idx int) bool {
	return ls.Type(idx) == api.LUA_TNONE
}

func (ls *luaState) IsNil(idx int) bool {
	return ls.Type(idx) == api.LUA_TNIL
}

func (ls *luaState) IsNoneOrNil(idx int) bool {
	return ls.Type(idx) <= api.LUA_TNIL
}

func (ls *luaState) IsBoolean(idx int) bool {
	return ls.Type(idx) == api.LUA_TBOOLEAN
}

func (ls *luaState) IsString(idx int) bool {
	t := ls.Type(idx)
	return t == api.LUA_TSTRING || t == api.LUA_TNUMBER
}

func (ls *luaState) IsNumber(idx int) bool {
	_, ok := ls.ToNumberX(idx)

	return ok
}

func (ls *luaState) IsInteger(idx int) bool {
	val := ls.stack.get(idx)
	_, ok := val.(int64)
	return ok
}

func (ls *luaState) ToBoolean(idx int) bool {
	val := ls.stack.get(idx)
	return convertToBoolean(val)
}

func convertToBoolean(val luaValue) bool {
	switch x := val.(type) {
	case nil:
		return false
	case bool:
		return x
	default:
		return true
	}
}

func (ls *luaState) ToNumber(idx int) float64 {
	n, _ := ls.ToNumberX(idx)
	return n
}

func (ls *luaState) ToNumberX(idx int) (float64, bool) {
	val := ls.stack.get(idx)
	return convertToFloat(val)
}

func (ls *luaState) ToInteger(idx int) int64 {
	i, _ := ls.ToIntegerX(idx)
	return i
}

func (ls *luaState) ToIntegerX(idx int) (int64, bool) {
	val := ls.stack.get(idx)
	return convertToInteger(val)
}

func (ls *luaState) ToStringX(idx int) (string, bool) {
	val := ls.stack.get(idx)
	switch x := val.(type) {
	case string:
		return x, true
	case int64, float64:
		s := fmt.Sprintf("%v", x)
		ls.stack.set(idx, s) // 注意这里会修改栈!
		return s, true
	default:
		return "", false
	}
}

func (ls *luaState) ToString(idx int) string {
	s, _ := ls.ToStringX(idx)
	return s
}

func (ls *luaState) IsGoFunction(idx int) bool {
	val := ls.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc != nil
	}
	return false
}

func (ls *luaState) ToGoFunction(idx int) api.GoFunction {
	val := ls.stack.get(idx)
	if c, ok := val.(*closure); ok {
		return c.goFunc
	}
	return nil
}
