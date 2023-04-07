package state

import (
	"fmt"
	. "lua_go/api"
	_ "lua_go/binchunk"
	"testing"
)

func TestArith(t *testing.T) {
	ls := New(20, nil)
	ls.PushInteger(1)
	ls.PushString("2.0")
	ls.PushString("3.0")
	ls.PushNumber(4.0)

	if stringifyStack(ls) != `[1]["2.0"]["3.0"][4]` {
		t.Fatalf(`stringifyStack(ls) != [1]["2.0"]["3.0"][4]`)
	}

	ls.Arith(LUA_OPADD)
	if stringifyStack(ls) != `[1]["2.0"][7]` {
		t.Fatalf(`stringifyStack(ls) != [1]["2.0"][7]`)
	}

	ls.Arith(LUA_OPBNOT)

	if stringifyStack(ls) != `[1]["2.0"][-8]` {
		t.Fatalf(`stringifyStack(ls) != [1]["2.0"][-8]`)
	}

	ls.Len(2)
	if stringifyStack(ls) != `[1]["2.0"][-8][3]` {
		t.Fatalf(`stringifyStack(ls) != [1]["2.0"][-8][3]`)
	}

	ls.Concat(3)
	if stringifyStack(ls) != `[1]["2.0-83"]` {
		t.Fatalf(`stringifyStack(ls) != [1]["2.0-83"]`)
	}
	ls.PushBoolean(ls.Compare(1, 2, LUA_OPEQ))
	if stringifyStack(ls) != `[1]["2.0-83"][false]` {
		t.Fatalf(`stringifyStack(ls) != [1]["2.0-83"][false]`)
	}
}

func stringifyStack(ls LuaState) string {
	var ans = ""
	top := ls.GetTop()
	for i := 1; i <= top; i++ {
		t := ls.Type(i)
		switch t {
		case LUA_TBOOLEAN:
			ans += fmt.Sprintf("[%t]", ls.ToBoolean(i))
		case LUA_TNUMBER:
			ans += fmt.Sprintf("[%g]", ls.ToNumber(i))
		case LUA_TSTRING:
			ans += fmt.Sprintf("[%q]", ls.ToString(i))
		default: // other values
			ans += fmt.Sprintf("[%s]", ls.TypeName(t))
		}
	}
	return ans
}
