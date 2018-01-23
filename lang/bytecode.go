package lang

import "fmt"

type bytecode struct {
	Instructions []Instr
}

func (b *bytecode) nextInstrPtr() Address {
	return Address(len(b.Instructions))
}

func (b *bytecode) write(instr Instr) Address {
	ip := b.nextInstrPtr()
	b.Instructions = append(b.Instructions, instr)
	return ip
}

func (b *bytecode) append(blob bytecode) Address {
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

func (b *bytecode) overwrite(addr Address, instr Instr) {
	b.Instructions[addr] = instr
}

func (b *bytecode) String() (out string) {
	for i, instr := range b.Instructions {
		if i > 0 {
			out += "\n"
		}
		out += fmt.Sprintf("%s %s", Address(i), instr)
	}
	return out
}

func sprintfArgs(name string, args ...interface{}) (out string) {
	if len(args) == 0 {
		return name
	}

	out = fmt.Sprintf("%-8s", name)
	for i := 0; i < len(args); i++ {
		if i < len(args)-1 {
			out += fmt.Sprintf("%-8v", args[i])
		} else {
			out += fmt.Sprintf("%v", args[i])
		}
	}
	return out
}

type Instr interface {
	fmt.Stringer
	isInstr()
}

type Address uint32

func (a Address) String() string {
	return fmt.Sprintf("0x%04x", uint32(a))
}

type InstrAddressed interface {
	Instr
	offset(Address) InstrAddressed
}

type InstrHalt struct{}

func (i InstrHalt) String() string { return "halt" }
func (i InstrHalt) isInstr()       {}

type InstrNOP struct{}

func (i InstrNOP) String() string { return "nop" }
func (i InstrNOP) isInstr()       {}

type InstrJump struct {
	addr Address
}

func (i InstrJump) String() string                       { return sprintfArgs("jmp", i.addr) }
func (i InstrJump) offset(offset Address) InstrAddressed { return InstrJump{i.addr + offset} }
func (i InstrJump) isInstr()                             {}

type InstrJumpTrue struct {
	addr Address
}

func (i InstrJumpTrue) String() string                       { return sprintfArgs("jmpt", i.addr) }
func (i InstrJumpTrue) offset(offset Address) InstrAddressed { return InstrJumpTrue{i.addr + offset} }
func (i InstrJumpTrue) isInstr()                             {}

type InstrJumpFalse struct {
	addr Address
}

func (i InstrJumpFalse) String() string                       { return sprintfArgs("jmpf", i.addr) }
func (i InstrJumpFalse) offset(offset Address) InstrAddressed { return InstrJumpFalse{i.addr + offset} }
func (i InstrJumpFalse) isInstr()                             {}

type InstrPush struct {
	Val Object
}

func (i InstrPush) String() string { return sprintfArgs("push", i.Val) }
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

func (i InstrReserve) String() string { return sprintfArgs("alloc", i.Name) }
func (i InstrReserve) isInstr()       {}

type InstrStore struct {
	Name   string
	Symbol *UniqueSymbol
}

func (i InstrStore) String() string { return sprintfArgs("store", i.Name) }
func (i InstrStore) isInstr()       {}

type InstrLoadSelf struct{}

func (i InstrLoadSelf) String() string { return "self" }
func (i InstrLoadSelf) isInstr()       {}

type InstrLoad struct {
	Name   string
	Symbol *UniqueSymbol
}

func (i InstrLoad) String() string { return sprintfArgs("load", i.Name) }
func (i InstrLoad) isInstr()       {}

type InstrDispatch struct {
	args int
}

func (i InstrDispatch) String() string { return sprintfArgs("call", i.args) }
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
