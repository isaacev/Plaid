package vm

import "fmt"

// Opcode uniquely identifies each type of instruction
type Opcode uint8

// Opcodes classify each type of instruction
const (
	Halt Opcode = iota
	Push
	Pop
	Store
	Add
	Print
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

// InstrPush adds its argument to the top of the VM expression stack
type InstrPush struct {
	Val Object
}

func (ip InstrPush) String() string { return fmt.Sprintf("Push %8s", ip.Val) }
func (ip InstrPush) isInstr()       {}

// InstrPop remove the top value from the stack and discard the value
type InstrPop struct{}

func (ip InstrPop) String() string { return "Pop" }
func (ip InstrPop) isInstr()       {}

// InstrStore remove the top value from the stack and store it in a register
type InstrStore struct {
	Reg *Register
}

func (is InstrStore) String() string { return fmt.Sprintf("Store %8s", is.Reg) }
func (is InstrStore) isInstr()       {}

// InstrAdd pops top 2 values from stack, adds them, pushes sum back onto stack
type InstrAdd struct{}

func (ia InstrAdd) String() string { return "Add" }
func (ia InstrAdd) isInstr()       {}

// InstrPrint pops top value from stack and prints it to STDOUT
type InstrPrint struct{}

func (ip InstrPrint) String() string { return "Print" }
func (ip InstrPrint) isInstr()       {}
