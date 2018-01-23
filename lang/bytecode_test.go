package lang

import "testing"

func TestInstrHalt(t *testing.T) {
	instr := InstrHalt{}
	instr.isInstr()
	expectString(t, instr.String(), "halt")
}

func TestInstrNOP(t *testing.T) {
	instr := InstrNOP{}
	instr.isInstr()
	expectString(t, instr.String(), "nop")
}

func TestInstrJump(t *testing.T) {
	instr := InstrJump{addr: 100}
	instr.isInstr()
	expectString(t, instr.String(), "jmp 0x0064")
}

func TestInstrJumpTrue(t *testing.T) {
	instr := InstrJumpTrue{addr: 100}
	instr.isInstr()
	expectString(t, instr.String(), "jmpt 0x0064")
}

func TestInstrJumpFalse(t *testing.T) {
	instr := InstrJumpFalse{addr: 100}
	instr.isInstr()
	expectString(t, instr.String(), "jmpf 0x0064")
}

func TestInstrPush(t *testing.T) {
	instr := InstrPush{ObjectStr{"abc"}}
	instr.isInstr()
	expectString(t, instr.String(), `push "abc"`)
}

func TestInstrPop(t *testing.T) {
	instr := InstrPop{}
	instr.isInstr()
	expectString(t, instr.String(), "pop")
}

func TestInstrCopy(t *testing.T) {
	instr := InstrCopy{}
	instr.isInstr()
	expectString(t, instr.String(), "copy")
}

func TestInstrReserve(t *testing.T) {
	instr := InstrReserve{Name: "foo"}
	instr.isInstr()
	expectString(t, instr.String(), "alloc foo")
}

func TestInstrStore(t *testing.T) {
	instr := InstrStore{Name: "foo"}
	instr.isInstr()
	expectString(t, instr.String(), "store foo")
}

func TestInstrLoadSelf(t *testing.T) {
	instr := InstrLoadSelf{}
	instr.isInstr()
	expectString(t, instr.String(), "self")
}

func TestInstrLoad(t *testing.T) {
	instr := InstrLoad{Name: "foo"}
	instr.isInstr()
	expectString(t, instr.String(), "load foo")
}

func TestInstrDispatch(t *testing.T) {
	instr := InstrDispatch{args: 5}
	instr.isInstr()
	expectString(t, instr.String(), "call 5")
}

func TestInstrNone(t *testing.T) {
	instr := InstrNone{}
	instr.isInstr()
	expectString(t, instr.String(), "none")
}

func TestInstrReturn(t *testing.T) {
	instr := InstrReturn{}
	instr.isInstr()
	expectString(t, instr.String(), "ret")
}

func TestInstrAdd(t *testing.T) {
	instr := InstrAdd{}
	instr.isInstr()
	expectString(t, instr.String(), "add")
}

func TestInstrSub(t *testing.T) {
	instr := InstrSub{}
	instr.isInstr()
	expectString(t, instr.String(), "sub")
}

func TestInstrMul(t *testing.T) {
	instr := InstrMul{}
	instr.isInstr()
	expectString(t, instr.String(), "mul")
}

func TestInstrLT(t *testing.T) {
	instr := InstrLT{}
	instr.isInstr()
	expectString(t, instr.String(), "cmplt")
}

func TestInstrLTEquals(t *testing.T) {
	instr := InstrLTEquals{}
	instr.isInstr()
	expectString(t, instr.String(), "cmplte")
}

func TestInstrGT(t *testing.T) {
	instr := InstrGT{}
	instr.isInstr()
	expectString(t, instr.String(), "cmpgt")
}

func TestInstrGTEquals(t *testing.T) {
	instr := InstrGTEquals{}
	instr.isInstr()
	expectString(t, instr.String(), "cmpgte")
}
