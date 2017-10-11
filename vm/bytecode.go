package vm

import (
	"fmt"
	"plaid/debug"
	"strings"
)

// Bytecode holds a list of instructions that can be executed by the VM
type Bytecode struct {
	Instrs   []Instr
	Children []*Bytecode
}

// Descend adds a new bytecode chunk in a child relationship to the current
// bytecode chunk. This is done for to ease the visualization of bytecode
// chunks during debugging
func (bc *Bytecode) Descend() *Bytecode {
	child := &Bytecode{}
	bc.Children = append(bc.Children, child)
	return child
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

// StringChildren satisfies the requirements of the debug.StringTree interface
// so that a scope tree can be pretty printed
func (bc *Bytecode) StringChildren() []debug.StringTree {
	var children []debug.StringTree
	for _, child := range bc.Children {
		children = append(children, child)
	}
	return children
}

func indentBlock(indent string, source string) string {
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}

	return strings.Join(lines, "\n")
}
