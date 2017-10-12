package vm

import (
	"fmt"
	"plaid/types"
)

// Module holds all the data necessary to build and evaluate a code module
// including all child closures and any dependency information
type Module struct {
	Name    string
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
// and bytecode data required to build a closure during execution. Each
// function literal is converted to a single ClosureTemplate but each template
// may be converted to any number of Closures during runtime
type ClosureTemplate struct {
	ID         int
	Parameters []*CellTemplate
	Bytecode   *Bytecode
}

func (c *ClosureTemplate) String() string {
	return fmt.Sprintf("<closure template #%d>", c.ID)
}

func (c *ClosureTemplate) isObject() {}

var uniqueClosureID int

// MakeClosureTemplate is a helper function to create a closure template from
// a given set of parameter CellTemplates and a Bytecode blob. This function
// should be called during codegen exactly once for each function literal
func MakeClosureTemplate(params []*CellTemplate, bc *Bytecode) *ClosureTemplate {
	nextID := uniqueClosureID
	uniqueClosureID++
	return &ClosureTemplate{
		ID:         nextID,
		Parameters: params,
		Bytecode:   bc,
	}
}

// Closure is an object created during runtime from a ClosureTemplate that binds
// a runtime environment to a Bytecode blob so that any evaluation of that
// Bytecode is done in the context of the bound environment. Closures are
// objects that can be passed, referenced, and manipulated the same as any other
// object in the virtual machine.
type Closure struct {
	Env        *Env
	Parameters []*CellTemplate
	Bytecode   *Bytecode
}

func (c *Closure) String() string { return "<closure>" }
func (c *Closure) isObject()      {}
