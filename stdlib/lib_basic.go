package stdlib

import (
	"fmt"
	"lua_go/api"
	"strconv"
	"strings"
)

var baseFuncs = map[string]api.GoFunction{
	"print":        basePrint,
	"assert":       baseAssert,
	"error":        baseError,
	"select":       baseSelect,
	"ipairs":       baseIPairs,
	"pairs":        basePairs,
	"next":         baseNext,
	"load":         baseLoad,
	"loadfile":     baseLoadFile,
	"dofile":       baseDoFile,
	"pcall":        basePCall,
	"xpcall":       baseXPCall,
	"getmetatable": baseGetMetatable,
	"setmetatable": baseSetMetatable,
	"rawequal":     baseRawEqual,
	"rawlen":       baseRawLen,
	"rawget":       baseRawGet,
	"rawset":       baseRawSet,
	"type":         baseType,
	"tostring":     baseToString,
	"tonumber":     baseToNumber,
	/* placeholders */
	"_G":       nil,
	"_VERSION": nil,
}

func OpenBaseLib(ls api.LuaState) int {
	/* open lib into global table */
	ls.PushGlobalTable()
	ls.SetFuncs(baseFuncs, 0)
	/* set global _G */
	ls.PushValue(-1)
	ls.SetField(-2, "_G")
	/* set global _VERSION */
	ls.PushString("Lua 5.3") // todo
	ls.SetField(-2, "_VERSION")
	return 1
}

func basePrint(ls api.LuaState) int {
	n := ls.GetTop() /* number of arguments */
	ls.GetGlobal("tostring")
	for i := 1; i <= n; i++ {
		ls.PushValue(-1) /* function to be called */
		ls.PushValue(i)  /* value to print */
		ls.Call(1, 1)
		s, ok := ls.ToStringX(-1) /* get result */
		if !ok {
			return ls.Error2("'tostring' must return a string to 'print'")
		}
		if i > 1 {
			fmt.Print("\t")
		}
		fmt.Print(s)
		ls.Pop(1) /* pop result */
	}
	fmt.Println()
	return 0
}

func baseAssert(ls api.LuaState) int {
	if ls.ToBoolean(1) { /* condition is true */
		return ls.GetTop() /* return all arguments */
	} else { /* error */
		ls.CheckAny(1)                     /* there must be a condition */
		ls.Remove(1)                       /* remove it */
		ls.PushString("assertion failed!") /* default message */
		ls.SetTop(1)                       /* leave only message (default if no other one) */
		return baseError(ls)               /* call 'error' */
	}
}

func baseError(ls api.LuaState) int {
	// level := int(ls.OptInteger(2, 1))
	ls.SetTop(1)
	// if ls.Type(1) == api.LUA_TSTRING && level > 0 {
	// 	ls.Where(level) /* add extra information */
	// 	ls.PushValue(1)
	// 	ls.Concat(2)
	// }
	return ls.Error()
}

func baseSelect(ls api.LuaState) int {
	n := int64(ls.GetTop())
	if ls.Type(1) == api.LUA_TSTRING && ls.CheckString(1) == "#" {
		ls.PushInteger(n - 1)
		return 1
	} else {
		i := ls.CheckInteger(1)
		if i < 0 {
			i = n + i
		} else if i > n {
			i = n
		}
		ls.ArgCheck(1 <= i, 1, "index out of range")
		return int(n - i)
	}
}

func baseIPairs(ls api.LuaState) int {
	ls.CheckAny(1)
	ls.PushGoFunction(iPairsAux) /* iteration function */
	ls.PushValue(1)              /* state */
	ls.PushInteger(0)            /* initial value */
	return 3
}

func _iPairsAux(ls api.LuaState) int {
	i := ls.ToInteger(2) + 1
	ls.PushInteger(i)
	if ls.GetI(1, i) == api.LUA_TNIL {
		return 1
	} else {
		return 2
	}
}

func iPairsAux(ls api.LuaState) int {
	i := ls.CheckInteger(2) + 1
	ls.PushInteger(i)
	if ls.GetI(1, i) == api.LUA_TNIL {
		return 1
	} else {
		return 2
	}
}

func basePairs(ls api.LuaState) int {
	ls.CheckAny(1)
	if ls.GetMetafield(1, "__pairs") == api.LUA_TNIL { /* no metamethod? */
		ls.PushGoFunction(baseNext) /* will return generator, */
		ls.PushValue(1)             /* state, */
		ls.PushNil()
	} else {
		ls.PushValue(1) /* argument 'self' to metamethod */
		ls.Call(1, 3)   /* get 3 values from metamethod */
	}
	return 3
}

// next (table [, index])
// http://www.lua.org/manual/5.3/manual.html#pdf-next
// lua-5.3.4/src/lbaselib.c#luaB_next()
func baseNext(ls api.LuaState) int {
	ls.CheckType(1, api.LUA_TTABLE)
	ls.SetTop(2) /* create a 2nd argument if there isn't one */
	if ls.Next(1) {
		return 2
	} else {
		ls.PushNil()
		return 1
	}
}

// load (chunk [, chunkname [, mode [, env]]])
// http://www.lua.org/manual/5.3/manual.html#pdf-load
// lua-5.3.4/src/lbaselib.c#luaB_load()
func baseLoad(ls api.LuaState) int {
	var status int
	chunk, isStr := ls.ToStringX(1)
	mode := ls.OptString(3, "bt")
	env := 0 /* 'env' index or 0 if no 'env' */
	if !ls.IsNone(4) {
		env = 4
	}
	if isStr { /* loading a string? */
		chunkname := ls.OptString(2, chunk)
		status = ls.Load([]byte(chunk), chunkname, mode)
	} else { /* loading from a reader function */
		panic("loading from a reader function") // todo
	}
	return loadAux(ls, status, env)
}

// lua-5.3.4/src/lbaselib.c#load_aux()
func loadAux(ls api.LuaState, status, envIdx int) int {
	if status == api.LUA_OK {
		if envIdx != 0 { /* 'env' parameter? */
			panic("todo!")
		}
		return 1
	} else { /* error (message is on top of the stack) */
		ls.PushNil()
		ls.Insert(-2) /* put before error message */
		return 2      /* return nil plus error message */
	}
}

// loadfile ([filename [, mode [, env]]])
// http://www.lua.org/manual/5.3/manual.html#pdf-loadfile
// lua-5.3.4/src/lbaselib.c#luaB_loadfile()
func baseLoadFile(ls api.LuaState) int {
	fname := ls.OptString(1, "")
	mode := ls.OptString(1, "bt")
	env := 0 /* 'env' index or 0 if no 'env' */
	if !ls.IsNone(3) {
		env = 3
	}
	status := ls.LoadFileX(fname, mode)
	return loadAux(ls, status, env)
}

// dofile ([filename])
// http://www.lua.org/manual/5.3/manual.html#pdf-dofile
// lua-5.3.4/src/lbaselib.c#luaB_dofile()
func baseDoFile(ls api.LuaState) int {
	fname := ls.OptString(1, "bt")
	ls.SetTop(1)
	if ls.LoadFile(fname) != api.LUA_OK {
		return ls.Error()
	}
	ls.Call(0, api.LUA_MULTRET)
	return ls.GetTop() - 1
}

// pcall (f [, arg1, ···])
// http://www.lua.org/manual/5.3/manual.html#pdf-pcall
func basePCall(ls api.LuaState) int {
	nArgs := ls.GetTop() - 1
	status := ls.PCall(nArgs, -1, 0)
	ls.PushBoolean(status == api.LUA_OK)
	ls.Insert(1)
	return ls.GetTop()
}

// xpcall (f, msgh [, arg1, ···])
// http://www.lua.org/manual/5.3/manual.html#pdf-xpcall
func baseXPCall(ls api.LuaState) int {
	panic("todo!")
}

// getmetatable (object)
// http://www.lua.org/manual/5.3/manual.html#pdf-getmetatable
// lua-5.3.4/src/lbaselib.c#luaB_getmetatable()
func baseGetMetatable(ls api.LuaState) int {
	ls.CheckAny(1)
	if !ls.GetMetatable(1) {
		ls.PushNil()
		return 1 /* no metatable */
	}
	ls.GetMetafield(1, "__metatable")
	return 1 /* returns either __metatable field (if present) or metatable */

}

// setmetatable (table, metatable)
// http://www.lua.org/manual/5.3/manual.html#pdf-setmetatable
// lua-5.3.4/src/lbaselib.c#luaB_setmetatable()
func baseSetMetatable(ls api.LuaState) int {
	t := ls.Type(2)
	ls.CheckType(1, api.LUA_TTABLE)
	ls.ArgCheck(t == api.LUA_TNIL || t == api.LUA_TTABLE, 2,
		"nil or table expected")
	if ls.GetMetafield(1, "__metatable") != api.LUA_TNIL {
		return ls.Error2("cannot change a protected metatable")
	}
	ls.SetTop(2)
	ls.SetMetatable(1)
	return 1
}

// rawequal (v1, v2)
// http://www.lua.org/manual/5.3/manual.html#pdf-rawequal
// lua-5.3.4/src/lbaselib.c#luaB_rawequal()
func baseRawEqual(ls api.LuaState) int {
	ls.CheckAny(1)
	ls.CheckAny(2)
	ls.PushBoolean(ls.RawEqual(1, 2))
	return 1
}

// rawlen (v)
// http://www.lua.org/manual/5.3/manual.html#pdf-rawlen
// lua-5.3.4/src/lbaselib.c#luaB_rawlen()
func baseRawLen(ls api.LuaState) int {
	t := ls.Type(1)
	ls.ArgCheck(t == api.LUA_TTABLE || t == api.LUA_TSTRING, 1,
		"table or string expected")
	ls.PushInteger(int64(ls.RawLen(1)))
	return 1
}

// rawget (table, index)
// http://www.lua.org/manual/5.3/manual.html#pdf-rawget
// lua-5.3.4/src/lbaselib.c#luaB_rawget()
func baseRawGet(ls api.LuaState) int {
	ls.CheckType(1, api.LUA_TTABLE)
	ls.CheckAny(2)
	ls.SetTop(2)
	ls.RawGet(1)
	return 1
}

// rawset (table, index, value)
// http://www.lua.org/manual/5.3/manual.html#pdf-rawset
// lua-5.3.4/src/lbaselib.c#luaB_rawset()
func baseRawSet(ls api.LuaState) int {
	ls.CheckType(1, api.LUA_TTABLE)
	ls.CheckAny(2)
	ls.CheckAny(3)
	ls.SetTop(3)
	ls.RawSet(1)
	return 1
}

// type (v)
// http://www.lua.org/manual/5.3/manual.html#pdf-type
// lua-5.3.4/src/lbaselib.c#luaB_type()
func baseType(ls api.LuaState) int {
	t := ls.Type(1)
	ls.ArgCheck(t != api.LUA_TNONE, 1, "value expected")
	ls.PushString(ls.TypeName(t))
	return 1
}

// tostring (v)
// http://www.lua.org/manual/5.3/manual.html#pdf-tostring
// lua-5.3.4/src/lbaselib.c#luaB_tostring()
func baseToString(ls api.LuaState) int {
	ls.CheckAny(1)
	ls.ToString2(1)
	return 1
}

// tonumber (e [, base])
// http://www.lua.org/manual/5.3/manual.html#pdf-tonumber
// lua-5.3.4/src/lbaselib.c#luaB_tonumber()
func baseToNumber(ls api.LuaState) int {
	if ls.IsNoneOrNil(2) { /* standard conversion? */
		ls.CheckAny(1)
		if ls.Type(1) == api.LUA_TNUMBER { /* already a number? */
			ls.SetTop(1) /* yes; return it */
			return 1
		} else {
			if s, ok := ls.ToStringX(1); ok {
				if ls.StringToNumber(s) {
					return 1 /* successful conversion to number */
				} /* else not a number */
			}
		}
	} else {
		ls.CheckType(1, api.LUA_TSTRING) /* no numbers as strings */
		s := strings.TrimSpace(ls.ToString(1))
		base := int(ls.CheckInteger(2))
		ls.ArgCheck(2 <= base && base <= 36, 2, "base out of range")
		if n, err := strconv.ParseInt(s, base, 64); err == nil {
			ls.PushInteger(n)
			return 1
		} /* else not a number */
	} /* else not a number */
	ls.PushNil() /* not a number */
	return 1
}
