package vm

import (
	"fmt"
)

// Bytecode holds a list of instructions that can be executed by the VM
type Bytecode struct {
	Instrs  []Instr
	Closure *ClosureTemplate
}

func (bc *Bytecode) Write(instr Instr) uint32 {
	ip := bc.NextIP()
	bc.Instrs = append(bc.Instrs, instr)
	return ip
}

// Overwrite clobbers a previously written instruction
func (bc *Bytecode) Overwrite(offset uint32, instr Instr) {
	bc.Instrs[offset] = instr
}

// NextIP returns the offset of the next instruction to be written
func (bc *Bytecode) NextIP() uint32 {
	return uint32(len(bc.Instrs))
}

func (bc *Bytecode) String() (out string) {
	for o, instr := range bc.Instrs {
		if o > 0 {
			out += "\n"
		}
		out += fmt.Sprintf("%04d %s", o, instr)
	}
	return out
}
