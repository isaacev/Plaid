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

func (ih InstrHalt) String() string { return "Halt" }
func (ih InstrHalt) isInstr()       {}

// InstrNOP is a non-operation instruction that does nothing
type InstrNOP struct{}

func (nop InstrNOP) String() string { return "NOP" }
func (nop InstrNOP) isInstr()       {}

// InstrPush adds its argument to the top of the VM expression stack
type InstrPush struct {
	Val Object
}

func (ip InstrPush) String() string { return fmt.Sprintf("Push\t%s", ip.Val) }
func (ip InstrPush) isInstr()       {}

// InstrPop remove the top value from the stack and discard the value
type InstrPop struct{}

func (ip InstrPop) String() string { return "Pop" }
func (ip InstrPop) isInstr()       {}

// InstrCopy duplicates the top value from the stack and pushes it onto the stack
type InstrCopy struct{}

func (ic InstrCopy) String() string { return "Copy" }
func (ic InstrCopy) isInstr()       {}

// InstrReserve allocates registers for local variables
type InstrReserve struct {
	Template *CellTemplate
}

func (ir InstrReserve) String() string { return fmt.Sprintf("Reserve\t%s", ir.Template) }
func (ir InstrReserve) isInstr()       {}

// InstrStore remove the top value from the stack and store it in a register
type InstrStore struct {
	Template *CellTemplate
}

func (is InstrStore) String() string { return fmt.Sprintf("Store\t%s", is.Template) }
func (is InstrStore) isInstr()       {}

// InstrLoad reads a register and pushes its contents onto the stack
type InstrLoad struct {
	Template *CellTemplate
}

func (il InstrLoad) String() string { return fmt.Sprintf("Load\t%s", il.Template) }
func (il InstrLoad) isInstr()       {}

// InstrDispatch reads arguments from the stack and passes them to the callee
type InstrDispatch struct {
	NumArgs int
}

func (id InstrDispatch) String() string { return fmt.Sprintf("Dispatch\t%d", id.NumArgs) }
func (id InstrDispatch) isInstr()       {}

// InstrNone adds a nothing object to the stack to help handling void
// functions that return no values
type InstrNone struct{}

func (in InstrNone) String() string { return fmt.Sprintf("PushNone") }
func (in InstrNone) isInstr()       {}

// InstrReturn exits the current function
type InstrReturn struct{}

func (ir InstrReturn) String() string { return fmt.Sprintf("Return") }
func (ir InstrReturn) isInstr()       {}

// InstrAdd pops top 2 values from stack, adds them, pushes sum back onto stack
type InstrAdd struct{}

func (ia InstrAdd) String() string { return "Add" }
func (ia InstrAdd) isInstr()       {}

// InstrSub pops top 2 values from stack, subtracts them, pushes difference back onto stack
type InstrSub struct{}

func (is InstrSub) String() string { return "Sub" }
func (is InstrSub) isInstr()       {}

// InstrPrint pops top value from stack and prints it to STDOUT
type InstrPrint struct{}

func (ip InstrPrint) String() string { return "Print" }
func (ip InstrPrint) isInstr()       {}
