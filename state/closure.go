package state

import (
	"lua_go/api"
	"lua_go/binchunk"
)

type closure struct {
	proto  *binchunk.Prototype // lua closure
	goFunc api.GoFunction      // go closure
}

func newLuaClosure(proto *binchunk.Prototype) *closure {
	return &closure{proto: proto}
}

func newGoClosure(f api.GoFunction) *closure {
	return &closure{goFunc: f}
}
