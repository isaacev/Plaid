package vm

import (
	"fmt"
	"plaid/types"
)

// Module holds all the data necessary to build and evaluate a code module
// including all child closures and any dependency information
type Module struct {
	Root    *ClosureTemplate
	Exports map[string]*Export
}

// Export describes an object made available to other modules. That object is
// described by a type for use during the type-checking stage of whatever
// modules use this export
type Export struct {
	Type   types.Type
	Object Object
}

// ClosureTemplate is generated in the codegen stage and encapsulates the scope
// and bytecode data required to build a closure during execution
type ClosureTemplate struct {
	ID         int
	Parameters []*CellTemplate
	Bytecode   *Bytecode
}

func (ct *ClosureTemplate) String() string { return fmt.Sprintf("<closure template #%d>", ct.ID) }
func (ct *ClosureTemplate) isObject()      {}

var uniqueClosureID int

// MakeClosureTemplate builds a closure template and assigns it a unique ID
func MakeClosureTemplate(params []*CellTemplate, bc *Bytecode) *ClosureTemplate {
	nextID := uniqueClosureID
	uniqueClosureID++
	return &ClosureTemplate{
		ID:         nextID,
		Parameters: params,
		Bytecode:   bc,
	}
}

// Closure is bytecode bound to a lexical scope
type Closure struct {
	Env        *Env
	Parameters []*CellTemplate
	Bytecode   *Bytecode
}

func (c *Closure) String() string { return "<closure>" }
func (c *Closure) isObject()      {}
