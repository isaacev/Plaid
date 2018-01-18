package lang

import "fmt"

type bytecode struct {
	Instructions []Instr
}

func (b *bytecode) nextInstrPtr() uint32 {
	return uint32(len(b.Instructions))
}

func (b *bytecode) write(instr Instr) uint32 {
	ip := b.nextInstrPtr()
	b.Instructions = append(b.Instructions, instr)
	return ip
}

func (b *bytecode) append(blob bytecode) uint32 {
	offset := b.nextInstrPtr()
	for _, instr := range blob.Instructions {
		if jump, ok := instr.(InstrAddressed); ok {
			b.Instructions = append(b.Instructions, jump.offset(offset))
		} else {
			b.Instructions = append(b.Instructions, instr)
		}
	}
	return b.nextInstrPtr()
}

func (b *bytecode) overwrite(addr uint32, instr Instr) {
	b.Instructions[addr] = instr
}

func (b *bytecode) String() (out string) {
	for i, instr := range b.Instructions {
		if i > 0 {
			out += "\n"
		}
		out += fmt.Sprintf("%04d %s", i, instr)
	}
	return out
}

type Instr interface {
	fmt.Stringer
	isInstr()
}

type InstrAddressed interface {
	Instr
	offset(uint32) InstrAddressed
}

type InstrHalt struct{}

func (i InstrHalt) String() string { return "halt" }
func (i InstrHalt) isInstr()       {}

type InstrNOP struct{}

func (i InstrNOP) String() string { return "nop" }
func (i InstrNOP) isInstr()       {}

type InstrJump struct {
	addr uint32
}

func (i InstrJump) String() string                      { return fmt.Sprintf("jmp %d", i.addr) }
func (i InstrJump) offset(offset uint32) InstrAddressed { return InstrJump{i.addr + offset} }
func (i InstrJump) isInstr()                            {}

type InstrJumpTrue struct {
	addr uint32
}

func (i InstrJumpTrue) String() string                      { return fmt.Sprintf("jmpt %d", i.addr) }
func (i InstrJumpTrue) offset(offset uint32) InstrAddressed { return InstrJumpTrue{i.addr + offset} }
func (i InstrJumpTrue) isInstr()                            {}

type InstrJumpFalse struct {
	addr uint32
}

func (i InstrJumpFalse) String() string                      { return fmt.Sprintf("jmpf %d", i.addr) }
func (i InstrJumpFalse) offset(offset uint32) InstrAddressed { return InstrJumpFalse{i.addr + offset} }
func (i InstrJumpFalse) isInstr()                            {}

type InstrPush struct {
	Val Object
}

func (i InstrPush) String() string { return fmt.Sprintf("push %s", i.Val.String()) }
func (i InstrPush) isInstr()       {}

type InstrPop struct{}

func (i InstrPop) String() string { return "pop" }
func (i InstrPop) isInstr()       {}

type InstrCopy struct{}

func (i InstrCopy) String() string { return "copy" }
func (i InstrCopy) isInstr()       {}

type InstrReserve struct {
	Name   string
	Symbol *UniqueSymbol
}

func (i InstrReserve) String() string { return fmt.Sprintf("alloc %s", i.Name) }
func (i InstrReserve) isInstr()       {}

type InstrStore struct {
	Name   string
	Symbol *UniqueSymbol
}

func (i InstrStore) String() string { return fmt.Sprintf("store %s", i.Name) }
func (i InstrStore) isInstr()       {}

type InstrLoadSelf struct{}

func (i InstrLoadSelf) String() string { return "self" }
func (i InstrLoadSelf) isInstr()       {}

type InstrLoad struct {
	Name   string
	Symbol *UniqueSymbol
}

func (i InstrLoad) String() string { return fmt.Sprintf("load %s", i.Name) }
func (i InstrLoad) isInstr()       {}

type InstrDispatch struct {
	args int
}

func (i InstrDispatch) String() string { return fmt.Sprintf("call %d", i.args) }
func (i InstrDispatch) isInstr()       {}

type InstrNone struct{}

func (i InstrNone) String() string { return "none" }
func (i InstrNone) isInstr()       {}

type InstrReturn struct{}

func (i InstrReturn) String() string { return "ret" }
func (i InstrReturn) isInstr()       {}

type InstrAdd struct{}

func (i InstrAdd) String() string { return "add" }
func (i InstrAdd) isInstr()       {}

type InstrSub struct{}

func (i InstrSub) String() string { return "sub" }
func (i InstrSub) isInstr()       {}

type InstrMul struct{}

func (i InstrMul) String() string { return "mul" }
func (i InstrMul) isInstr()       {}

type InstrLT struct{}

func (i InstrLT) String() string { return "cmplt" }
func (i InstrLT) isInstr()       {}

type InstrLTEquals struct{}

func (i InstrLTEquals) String() string { return "cmplte" }
func (i InstrLTEquals) isInstr()       {}

type InstrGT struct{}

func (i InstrGT) String() string { return "cmpgt" }
func (i InstrGT) isInstr()       {}

type InstrGTEquals struct{}

func (i InstrGTEquals) String() string { return "cmpgte" }
func (i InstrGTEquals) isInstr()       {}

type InstrPrint struct{}

func (i InstrPrint) String() string { return "print" }
func (i InstrPrint) isInstr()       {}
