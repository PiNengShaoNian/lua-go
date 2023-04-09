package state

import "lua_go/api"

func (ls *luaState) Compare(idx1, idx2 int, op api.CompareOp) bool {
	a := ls.stack.get(idx1)
	b := ls.stack.get(idx2)
	switch op {
	case api.LUA_OPEQ:
		return _eq(a, b, ls)
	case api.LUA_OPLT:
		return _lt(a, b, ls)
	case api.LUA_OPLE:
		return _le(a, b, ls)
	default:
		panic("invalid compare op!")
	}
}

func _eq(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case *luaTable:
		if y, ok := b.(*luaTable); ok && x != y && ls != nil {
			if result, ok := callMetamethod(x, y, "__eq", ls); ok {
				return convertToBoolean(result)
			}
		}
		return a == b
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case string:
		y, ok := b.(string)
		return ok && x == y
	case int64:
		switch y := b.(type) {
		case int64:
			return x == y
		case float64:
			return float64(x) == y
		default:
			return false
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x == y
		case int64:
			return x == float64(y)
		default:
			return false
		}
	default:
		return a == b
	}
}

func _lt(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case string:
		if y, ok := b.(string); ok {
			return x < y
		}
	case int64:
		switch y := b.(type) {
		case int64:
			return x < y
		case float64:
			return float64(x) < y
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x < y
		case int64:
			return x < float64(y)
		}
	}

	if result, ok := callMetamethod(a, b, "__lt", ls); ok {
		return convertToBoolean(result)
	}
	panic("comparison error!")
}

func _le(a, b luaValue, ls *luaState) bool {
	switch x := a.(type) {
	case string:
		if y, ok := b.(string); ok {
			return x <= y
		}
	case int64:
		switch y := b.(type) {
		case int64:
			return x <= y
		case float64:
			return float64(x) <= y
		}
	case float64:
		switch y := b.(type) {
		case float64:
			return x <= y
		case int64:
			return x <= float64(y)
		}
	}

	if result, ok := callMetamethod(a, b, "__le", ls); ok {
		return convertToBoolean(result)
	} else if result, ok := callMetamethod(b, a, "__lt", ls); ok {
		return !convertToBoolean(result)
	} else {
		panic("comparison error!")
	}
}

func (ls *luaState) RawEqual(idx1, idx2 int) bool {
	if !ls.stack.isValid(idx1) || !ls.stack.isValid(idx2) {
		return false
	}

	a := ls.stack.get(idx1)
	b := ls.stack.get(idx2)
	return _eq(a, b, nil)
}
