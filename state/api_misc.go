package state

func (ls *luaState) Len(idx int) {
	val := ls.stack.get(idx)
	if s, ok := val.(string); ok {
		ls.stack.push(int64(len(s)))
	} else {
		panic("length error!")
	}
}

func (ls *luaState) Concat(n int) {
	if n == 0 {
		ls.stack.push("")
	} else if n >= 2 {
		for i := 1; i < n; i++ {
			if ls.IsString(-1) && ls.IsString(-2) {
				s2 := ls.ToString(-1)
				s1 := ls.ToString(-2)
				ls.stack.pop()
				ls.stack.pop()
				ls.stack.push(s1 + s2)
				continue
			}
			panic("concatenation error!")
		}
	}
	// n == 1, do nothing
}
