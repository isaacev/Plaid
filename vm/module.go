package vm

// Module is a collection of functions
type Module struct {
	Main      *Bytecode
	Functions []*Bytecode
}
