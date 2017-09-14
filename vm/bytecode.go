package vm

// Bytecode holds a list of instructions that can be executed by the VM
type Bytecode struct {
	Instrs []Instr
}

func (bc *Bytecode) Write(instr Instr) {
	bc.Instrs = append(bc.Instrs, instr)
}
