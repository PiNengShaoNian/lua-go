package state

import (
	"fmt"
	"lua_go/api"
	"lua_go/stdlib"
	"os"
)

func (ls *luaState) Error2(format string, a ...interface{}) int {
	s := fmt.Sprintf(format, a...)
	ls.PushString(s)

	return ls.Error()
}

func (ls *luaState) ArgError(arg int, extraMsg string) int {
	// bad argument #arg to 'funcname' (extramsg)
	return ls.Error2("bad argument #%d (%s)", arg, extraMsg)
}

func (ls *luaState) CheckStack2(sz int, msg string) {
	if !ls.CheckStack(sz) {
		if msg != "" {
			ls.Error2("stack overflow (%s)", msg)
		} else {
			ls.Error2("stack overflow")
		}
	}
}

func (ls *luaState) ArgCheck(cond bool, arg int, extraMsg string) {
	if !cond {
		ls.ArgError(arg, extraMsg)
	}
}

func (ls *luaState) CheckAny(arg int) {
	if ls.Type(arg) == api.LUA_TNONE {
		ls.ArgError(arg, "value expected")
	}
}

func (ls *luaState) CheckType(arg int, t api.LuaType) {
	if ls.Type(arg) != t {
		ls.tagError(arg, t)
	}
}

func (ls *luaState) CheckInteger(arg int) int64 {
	i, ok := ls.ToIntegerX(arg)
	if !ok {
		ls.intError(arg)
	}
	return i
}

func (ls *luaState) CheckNumber(arg int) float64 {
	f, ok := ls.ToNumberX(arg)
	if !ok {
		ls.tagError(arg, api.LUA_TNUMBER)
	}
	return f
}

func (ls *luaState) CheckString(arg int) string {
	s, ok := ls.ToStringX(arg)
	if !ok {
		ls.tagError(arg, api.LUA_TSTRING)
	}

	return s
}

func (ls *luaState) OptInteger(arg int, def int64) int64 {
	if ls.IsNoneOrNil(arg) {
		return def
	}
	return ls.CheckInteger(arg)
}

func (ls *luaState) OptNumber(arg int, def float64) float64 {
	if ls.IsNoneOrNil(arg) {
		return def
	}
	return ls.CheckNumber(arg)
}

func (ls *luaState) OptString(arg int, def string) string {
	if ls.IsNoneOrNil(arg) {
		return def
	}
	return ls.CheckString(arg)
}

func (ls *luaState) DoFile(filename string) bool {
	return ls.LoadFile(filename) != api.LUA_OK ||
		ls.PCall(0, api.LUA_MULTRET, 0) != api.LUA_OK
}

func (ls *luaState) DoString(str string) bool {
	return ls.LoadString(str) != api.LUA_OK ||
		ls.PCall(0, api.LUA_MULTRET, 0) != api.LUA_OK
}

func (ls *luaState) LoadFile(filename string) int {
	return ls.LoadFileX(filename, "bt")
}

func (ls *luaState) LoadFileX(filename, mode string) int {
	if data, err := os.ReadFile(filename); err == nil {
		return ls.Load(data, "@"+filename, mode)
	}
	return api.LUA_ERRFILE
}

func (ls *luaState) LoadString(s string) int {
	return ls.Load([]byte(s), s, "bt")
}

func (ls *luaState) TypeName2(idx int) string {
	return ls.TypeName(ls.Type(idx))
}

func (ls *luaState) Len2(idx int) int64 {
	ls.Len(idx)
	i, isNum := ls.ToIntegerX(-1)
	if !isNum {
		ls.Error2("object length is not an integer")
	}
	ls.Pop(1)
	return i
}

func (ls *luaState) ToString2(idx int) string {
	if ls.CallMeta(idx, "__tostring") { /* metafield? */
		if !ls.IsString(-1) {
			ls.Error2("'__tostring' must return a string")
		}
	} else {
		switch ls.Type(idx) {
		case api.LUA_TNUMBER:
			if ls.IsInteger(idx) {
				ls.PushString(fmt.Sprintf("%d", ls.ToInteger(idx))) // todo
			} else {
				ls.PushString(fmt.Sprintf("%g", ls.ToNumber(idx))) // todo
			}
		case api.LUA_TSTRING:
			ls.PushValue(idx)
		case api.LUA_TBOOLEAN:
			if ls.ToBoolean(idx) {
				ls.PushString("true")
			} else {
				ls.PushString("false")
			}
		case api.LUA_TNIL:
			ls.PushString("nil")
		default:
			tt := ls.GetMetafield(idx, "__name") /* try name */
			var kind string
			if tt == api.LUA_TSTRING {
				kind = ls.CheckString(-1)
			} else {
				kind = ls.TypeName2(idx)
			}

			ls.PushString(fmt.Sprintf("%s: %p", kind, ls.ToPointer(idx)))
			if tt != api.LUA_TNIL {
				ls.Remove(-2) /* remove '__name' */
			}
		}
	}
	return ls.CheckString(-1)
}

func (ls *luaState) GetSubTable(idx int, fname string) bool {
	if ls.GetField(idx, fname) == api.LUA_TTABLE {
		return true /* table already there */
	}
	ls.Pop(1) /* remove previous result */
	idx = ls.stack.absIndex(idx)
	ls.NewTable()
	ls.PushValue(-1)        /* copy to be left at top */
	ls.SetField(idx, fname) /* assign new table to field */
	return false            /* false, because did not find table there */
}

func (ls *luaState) GetMetafield(obj int, event string) api.LuaType {
	if !ls.GetMetatable(obj) { /* no metatable? */
		return api.LUA_TNIL
	}

	ls.PushString(event)
	tt := ls.RawGet(-2)
	if tt == api.LUA_TNIL { /* is metafield nil? */
		ls.Pop(2) /* remove metatable and metafield */
	} else {
		ls.Remove(-2) /* remove only metatable */
	}
	return tt /* return metafield type */
}

func (ls *luaState) CallMeta(obj int, event string) bool {
	obj = ls.AbsIndex(obj)
	if ls.GetMetafield(obj, event) == api.LUA_TNIL { /* no metafield? */
		return false
	}

	ls.PushValue(obj)
	ls.Call(1, 1)
	return true
}

func (ls *luaState) OpenLibs() {
	libs := map[string]api.GoFunction{
		"_G":        stdlib.OpenBaseLib,
		"math":      stdlib.OpenMathLib,
		"table":     stdlib.OpenTableLib,
		"string":    stdlib.OpenStringLib,
		"utf8":      stdlib.OpenUTF8Lib,
		"os":        stdlib.OpenOSLib,
		"package":   stdlib.OpenPackageLib,
		"coroutine": stdlib.OpenCoroutineLib,
	}

	for name, fun := range libs {
		ls.RequireF(name, fun, true)
		ls.Pop(1)
	}
}

func (ls *luaState) RequireF(modname string, openf api.GoFunction, glb bool) {
	ls.GetSubTable(api.LUA_REGISTRYINDEX, "_LOADED")
	ls.GetField(-1, modname) /* LOADED[modname] */
	if !ls.ToBoolean(-1) {   /* package not already loaded? */
		ls.Pop(1) /* remove field */
		ls.PushGoFunction(openf)
		ls.PushString(modname)   /* argument to open function */
		ls.Call(1, 1)            /* call 'openf' to open module */
		ls.PushValue(-1)         /* make copy of module (call result) */
		ls.SetField(-3, modname) /* _LOADED[modname] = module */
	}
	ls.Remove(-2) /* remove _LOADED table */
	if glb {
		ls.PushValue(-1)      /* copy of module */
		ls.SetGlobal(modname) /* _G[modname] = module */
	}
}

func (ls *luaState) NewLib(l api.FuncReg) {
	ls.NewLibTable(l)
	ls.SetFuncs(l, 0)
}

func (ls *luaState) NewLibTable(l api.FuncReg) {
	ls.CreateTable(0, len(l))
}

func (ls *luaState) SetFuncs(l api.FuncReg, nup int) {
	ls.CheckStack2(nup, "too many upvalues")
	for name, fun := range l { /* fill the table with given functions */
		for i := 0; i < nup; i++ { /* copy upvalues to the top */
			ls.PushValue(-nup)
		}
		// r[-(nup+2)][name]=fun
		ls.PushGoClosure(fun, nup) /* closure with those upvalues */
		ls.SetField(-(nup + 2), name)
	}
	ls.Pop(nup) /* remove upvalues */
}

func (ls *luaState) intError(arg int) {
	if ls.IsNumber(arg) {
		ls.ArgError(arg, "number has no integer representation")
	} else {
		ls.tagError(arg, api.LUA_TNUMBER)
	}
}

func (ls *luaState) tagError(arg int, tag api.LuaType) {
	ls.typeError(arg, ls.TypeName(api.LuaType(tag)))
}

func (ls *luaState) typeError(arg int, tname string) int {
	var typeArg string /* name for the type of the actual argument */
	if ls.GetMetafield(arg, "__name") == api.LUA_TSTRING {
		typeArg = ls.ToString(-1) /* use the given type name */
	} else if ls.Type(arg) == api.LUA_TLIGHTUSERDATA {
		typeArg = "light userdata" /* special name for messages */
	} else {
		typeArg = ls.TypeName2(arg) /* standard name */
	}
	msg := tname + " expected, got " + typeArg
	ls.PushString(msg)
	return ls.ArgError(arg, msg)
}
