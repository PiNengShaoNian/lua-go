package state

import (
	"lua_go/number"
	"math"
)

type luaTable struct {
	metatable *luaTable
	arr       []luaValue
	_map      map[luaValue]luaValue
}

func newLuaTable(nArr, nRec int) *luaTable {
	t := &luaTable{}
	if nArr > 0 {
		t.arr = make([]luaValue, 0, nArr)
	}
	if nRec > 0 {
		t._map = make(map[luaValue]luaValue, nRec)
	}
	return t
}

func (lt *luaTable) get(key luaValue) luaValue {
	key = _floatToInteger(key)
	if idx, ok := key.(int64); ok {
		if idx >= 1 && idx <= int64(len(lt.arr)) {
			return lt.arr[idx-1]
		}
	}

	return lt._map[key]
}

func _floatToInteger(key luaValue) luaValue {
	if f, ok := key.(float64); ok {
		if i, ok := number.FloatToInteger(f); ok {
			return i
		}
	}
	return key
}

func (lt *luaTable) put(key, val luaValue) {
	if key == nil {
		panic("table index is nil!")
	}
	if f, ok := key.(float64); ok && math.IsNaN(f) {
		panic("table index is NaN!")
	}
	key = _floatToInteger(key)
	if idx, ok := key.(int64); ok && idx >= 1 {
		arrLen := int64(len(lt.arr))
		if idx <= arrLen {
			lt.arr[idx-1] = val
			if idx == arrLen && val == nil {
				lt._shrinkArray()
			}
			return
		}
		if idx == arrLen+1 {
			delete(lt._map, key)
			if val != nil {
				lt.arr = append(lt.arr, val)
				lt._expandArray()
			}
			return
		}
	}
	if val != nil {
		if lt._map == nil {
			lt._map = make(map[luaValue]luaValue, 8)
		}
		lt._map[key] = val
	} else {
		delete(lt._map, key)
	}
}

func (lt *luaTable) _shrinkArray() {
	for i := len(lt.arr) - 1; i >= 0; i-- {
		if lt.arr[i] == nil {
			lt.arr = lt.arr[0:i]
		}
	}
}

func (lt *luaTable) _expandArray() {
	for idx := int64(len(lt.arr)) + 1; true; idx++ {
		if val, found := lt._map[idx]; found {
			delete(lt._map, idx)
			lt.arr = append(lt.arr, val)
		} else {
			break
		}
	}
}

func (lt *luaTable) len() int {
	return len(lt.arr)
}

func (lt *luaTable) hasMetafield(fieldName string) bool {
	return lt.metatable != nil && lt.metatable.get(fieldName) != nil
}
