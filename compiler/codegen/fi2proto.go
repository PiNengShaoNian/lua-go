package codegen

import "lua_go/binchunk"

func toProto(fi *funcInfo) *binchunk.Prototype {
	proto := &binchunk.Prototype{
		NumParams:    byte(fi.numParams),
		MaxStackSize: byte(fi.maxRegs),
		Code:         fi.insts,
		Constants:    getConstants(fi),
		Upvalues:     getUpvalues(fi),
		Protos:       toProtos(fi.subFuncs),
		LineInfo:     []uint32{},
		LocVars:      []binchunk.LocVar{},
		UpvalueNames: []string{},
	}

	if fi.isVararg {
		proto.IsVararg = 1
	}

	return proto
}

func toProtos(fis []*funcInfo) []*binchunk.Prototype {
	protos := make([]*binchunk.Prototype, len(fis))

	for i, fi := range fis {
		protos[i] = toProto(fi)
	}

	return protos
}

func getConstants(fi *funcInfo) []interface{} {
	consts := make([]interface{}, len(fi.constants))
	for k, idx := range fi.constants {
		consts[idx] = k
	}
	return consts
}

func getUpvalues(fi *funcInfo) []binchunk.Upvalue {
	upvals := make([]binchunk.Upvalue, len(fi.upvalues))
	for _, uv := range fi.upvalues {
		if uv.locVarSlot >= 0 { // instack
			upvals[uv.index] = binchunk.Upvalue{Instack: 1, Idx: byte(uv.locVarSlot)}
		} else {
			upvals[uv.index] = binchunk.Upvalue{Instack: 0, Idx: byte(uv.upvalIndex)}
		}
	}
	return upvals
}
