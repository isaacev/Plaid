package vm

import (
	"fmt"
)

// Instr describes all bytecode instructions
type Instr interface {
	String() string
	isInstr()
}

// InstrHalt signals that the VM can stop executing
type InstrHalt struct{}

func (ih InstrHalt) String() string { return "halt" }
func (ih InstrHalt) isInstr()       {}

// InstrNOP is a non-operation instruction that does nothing
type InstrNOP struct{}

func (nop InstrNOP) String() string { return "nop" }
func (nop InstrNOP) isInstr()       {}

// InstrJump is a non-conditional jump
type InstrJump struct {
	IP uint32
}

func (ij InstrJump) String() string { return fmt.Sprintf("%-8s%04d", "jmp", ij.IP) }
func (ij InstrJump) isInstr()       {}

// InstrJumpTrue will jump if the top value on the stack is true
type InstrJumpTrue struct {
	IP uint32
}

func (ijc InstrJumpTrue) String() string { return fmt.Sprintf("%-8s%04d", "jmpt", ijc.IP) }
func (ijc InstrJumpTrue) isInstr()       {}

// InstrJumpFalse will jump if the top value on the stack is false
type InstrJumpFalse struct {
	IP uint32
}

func (ijc InstrJumpFalse) String() string { return fmt.Sprintf("%-8s%04d", "jmpf", ijc.IP) }
func (ijc InstrJumpFalse) isInstr()       {}

// InstrPush adds its argument to the top of the VM expression stack
type InstrPush struct {
	Val Object
}

func (ip InstrPush) String() string { return fmt.Sprintf("%-8s%s", "push", ip.Val) }
func (ip InstrPush) isInstr()       {}

// InstrPop remove the top value from the stack and discard the value
type InstrPop struct{}

func (ip InstrPop) String() string { return "pop" }
func (ip InstrPop) isInstr()       {}

// InstrCopy duplicates the top value from the stack and pushes it onto the stack
type InstrCopy struct{}

func (ic InstrCopy) String() string { return "copy" }
func (ic InstrCopy) isInstr()       {}

// InstrReserve allocates registers for local variables
type InstrReserve struct {
	Template *CellTemplate
}

func (ir InstrReserve) String() string { return fmt.Sprintf("%-8s%s", "alloc", ir.Template) }
func (ir InstrReserve) isInstr()       {}

// InstrStore remove the top value from the stack and store it in a register
type InstrStore struct {
	Template *CellTemplate
}

func (is InstrStore) String() string { return fmt.Sprintf("%-8s%s", "store", is.Template) }
func (is InstrStore) isInstr()       {}

// InstrLoadSelf pushes a copy of the current closure onto the stack so that
// it can be recursively called
type InstrLoadSelf struct{}

func (ils InstrLoadSelf) String() string { return "self" }
func (ils InstrLoadSelf) isInstr()       {}

// InstrLoad reads a register and pushes its contents onto the stack
type InstrLoad struct {
	Template *CellTemplate
}

func (il InstrLoad) String() string { return fmt.Sprintf("%-8s%s", "ld", il.Template) }
func (il InstrLoad) isInstr()       {}

// InstrDispatch reads arguments from the stack and passes them to the callee
type InstrDispatch struct {
	NumArgs int
}

func (id InstrDispatch) String() string { return fmt.Sprintf("%-8s%d", "call", id.NumArgs) }
func (id InstrDispatch) isInstr()       {}

// InstrNone adds a nothing object to the stack to help handling void
// functions that return no values
type InstrNone struct{}

func (in InstrNone) String() string { return "none" }
func (in InstrNone) isInstr()       {}

// InstrReturn exits the current function
type InstrReturn struct{}

func (ir InstrReturn) String() string { return "ret" }
func (ir InstrReturn) isInstr()       {}

// InstrAdd pops top 2 values from stack, adds them, pushes sum back onto stack
type InstrAdd struct{}

func (ia InstrAdd) String() string { return "add" }
func (ia InstrAdd) isInstr()       {}

// InstrSub pops top 2 values from stack, subtracts them, pushes difference back onto stack
type InstrSub struct{}

func (is InstrSub) String() string { return "sub" }
func (is InstrSub) isInstr()       {}

// InstrLT pops top 2 values from stack, pushes true if first is greater than second
type InstrLT struct{}

func (ilt InstrLT) String() string { return "cmplt" }
func (ilt InstrLT) isInstr()       {}

// InstrLTEquals pops top 2 values from stack, pushes true if first is greater than second
type InstrLTEquals struct{}

func (ilte InstrLTEquals) String() string { return "cmplte" }
func (ilte InstrLTEquals) isInstr()       {}

// InstrGT pops top 2 values from stack, pushes true if first is greater than second
type InstrGT struct{}

func (igt InstrGT) String() string { return "cmpgt" }
func (igt InstrGT) isInstr()       {}

// InstrGTEquals pops top 2 values from stack, pushes true if first is greater than second
type InstrGTEquals struct{}

func (igte InstrGTEquals) String() string { return "cmpgte" }
func (igte InstrGTEquals) isInstr()       {}
