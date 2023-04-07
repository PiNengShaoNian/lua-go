package vm

const MAXARG_Bx = 1<<18 - 1       // 262143
const MAXARG_sBx = MAXARG_Bx >> 1 // 131071

/*
31       22       13       5    0

	+-------+^------+-^-----+-^-----
	|b=9bits |c=9bits |a=8bits|op=6|
	+-------+^------+-^-----+-^-----
	|    bx=18bits    |a=8bits|op=6|
	+-------+^------+-^-----+-^-----
	|   sbx=18bits    |a=8bits|op=6|
	+-------+^------+-^-----+-^-----
	|    ax=26bits            |op=6|
	+-------+^------+-^-----+-^-----

31      23      15       7      0
*/
type Instruction uint32

func (ls Instruction) Opcode() int {
	return int(ls & 0x3F)
}

func (ls Instruction) ABC() (a, b, c int) {
	a = int(ls >> 6 & 0xFF)
	c = int(ls >> 14 & 0x1FF)
	b = int(ls >> 23 & 0x1FF)
	return
}

func (ls Instruction) ABx() (a, bx int) {
	a = int(ls >> 6 & 0xFF)
	bx = int(ls >> 14)
	return
}

func (ls Instruction) AsBx() (a, sbx int) {
	a, bx := ls.ABx()
	return a, bx - MAXARG_sBx
}

func (ls Instruction) Ax() int {
	return int(ls >> 6)
}

func (ls Instruction) OpName() string {
	return opcodes[ls.Opcode()].name
}

func (ls Instruction) OpMode() byte {
	return opcodes[ls.Opcode()].opMode
}

func (ls Instruction) BMode() byte {
	return opcodes[ls.Opcode()].argBMode
}

func (ls Instruction) CMode() byte {
	return opcodes[ls.Opcode()].argCMode
}
