package api

type LuaType = int
type ArithOp = int
type CompareOp = int

func LuaUpvalueIndex(i int) int {
	return LUA_REGISTRYINDEX - i
}

type BasicAPI interface {
	/* basic stack manipulation */
	GetTop() int
	AbsIndex(idx int) int
	CheckStack(n int) bool
	Pop(n int)
	Copy(fromIdx, toIdx int)
	PushValue(idx int)
	Replace(idx int)
	Insert(idx int)
	Remove(idx int)
	Rotate(idx, n int)
	SetTop(idx int)
	XMove(to LuaState, n int)
	/* access functions (stack -> GO) */
	TypeName(tp LuaType) string
	Type(idx int) LuaType
	IsNone(idx int) bool
	IsNil(idx int) bool
	IsNoneOrNil(idx int) bool
	IsBoolean(idx int) bool
	IsInteger(idx int) bool
	IsNumber(idx int) bool
	IsString(idx int) bool
	ToBoolean(idx int) bool
	ToInteger(idx int) int64
	ToIntegerX(idx int) (int64, bool)
	ToNumber(idx int) float64
	ToNumberX(idx int) (float64, bool)
	ToString(idx int) string
	ToStringX(idx int) (string, bool)
	ToThread(idx int) LuaState
	ToPointer(idx int) interface{}
	/* push functions (Go -> stack) */
	PushNil()
	PushBoolean(b bool)
	PushInteger(n int64)
	PushNumber(n float64)
	PushString(s string)
	Arith(op ArithOp)
	Compare(idx1, idx2 int, op CompareOp) bool
	Concat(n int)
	Len(idx int)
	/* get functions (Lua -> stack) */
	NewTable()
	CreateTable(nArr, nRec int)
	GetTable(idx int) LuaType
	GetField(idx int, k string) LuaType
	GetI(idx int, i int64) LuaType
	/* set functions (stack -> Lua) */
	SetTable(idx int)
	SetField(idx int, k string)
	SetI(idx int, i int64)
	Load(chunk []byte, chunkName, mode string) int
	Call(nArgs, nResults int)
	PushGoFunction(f GoFunction)
	IsGoFunction(idx int) bool
	ToGoFunction(idx int) GoFunction
	PushGlobalTable()
	PushThread() bool
	GetGlobal(name string) LuaType
	SetGlobal(name string)
	Register(name string, f GoFunction)
	PushGoClosure(f GoFunction, n int)
	GetMetatable(idx int) bool
	SetMetatable(idx int)
	RawLen(idx int) uint
	RawEqual(idx1, idx2 int) bool
	RawGet(idx int) LuaType
	RawSet(idx int)
	RawGetI(idx int, i int64) LuaType
	RawSetI(idx int, i int64)
	Next(idx int) bool
	Error() int
	PCall(nArgs, nResults, msgh int) int
	StringToNumber(s string) bool
	IsFunction(idx int) bool
	NewThread() LuaState
	Resume(from LuaState, nArgs int) int
	Yield(nResults int) int
	Status() int
	IsYieldable() bool
	GetStack() bool // debug
}

type LuaState interface {
	BasicAPI
	AuxLib
}

type GoFunction func(LuaState) int
