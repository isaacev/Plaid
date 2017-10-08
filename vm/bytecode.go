package vm

import (
	"fmt"
	"strings"
)

// Bytecode holds a list of instructions that can be executed by the VM
type Bytecode struct {
	Instrs     []Instr
	ChildFuncs []*Bytecode
}

func (bc *Bytecode) Write(instr Instr) {
	bc.Instrs = append(bc.Instrs, instr)
}

func (bc *Bytecode) String() string {
	out := ""
	for o, instr := range bc.Instrs {
		if o > 0 {
			out += "\n"
		}

		out += fmt.Sprintf("%04d %s", o, instr)
	}

	for _, child := range bc.ChildFuncs {
		out += "\n\n##"
		out += "\n## Child Function"
		out += "\n##\n"
		out += child.String()
	}

	return out
}

func indentBlock(indent string, source string) string {
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}

	return strings.Join(lines, "\n")
}
