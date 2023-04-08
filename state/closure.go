package state

import "lua_go/binchunk"

type closure struct {
	proto *binchunk.Prototype
}

func newLuaClosure(proto *binchunk.Prototype) *closure {
	return &closure{proto: proto}
}
