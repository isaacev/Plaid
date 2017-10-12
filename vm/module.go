package vm

import (
	"fmt"
	"plaid/debug"
	"plaid/types"
)

// Module holds all the data necessary to build and evaluate a code module
// including all child closures and any dependency information
type Module struct {
	Name    string
	Root    *ClosureTemplate
	Exports map[string]*Export
}

// HasExport returns true if this module has an export with the given name
func (m *Module) HasExport(name string) bool {
	if _, exists := m.Exports[name]; exists {
		return true
	}

	return false
}

// GetExport returns an Export struct if the given variable is exported, returns
// nil otherwise
func (m *Module) GetExport(name string) *Export {
	if m.HasExport(name) {
		return m.Exports[name]
	}

	return nil
}

// Export describes an object made available to other modules. That object is
// described by a type for use during the type-checking stage of whatever
// modules use this export
type Export struct {
	Type     types.Type
	Register *RegisterTemplate
	Object   Object
}

// ClosureTemplate is generated in the codegen stage and encapsulates the scope
// and bytecode data required to build a closure during execution. Each
// function literal is converted to a single ClosureTemplate but each template
// may be converted to any number of Closures during runtime
type ClosureTemplate struct {
	ID         int
	Parameters []*RegisterTemplate
	Bytecode   *Bytecode
	Enclosed   []*ClosureTemplate
}

// Enclose does some stuff
func (c *ClosureTemplate) Enclose(enclosed *ClosureTemplate) {
	c.Enclosed = append(c.Enclosed, enclosed)
}

func (c *ClosureTemplate) StringName() string {
	return fmt.Sprintf("<closure template #%d>", c.ID)
}

func (c *ClosureTemplate) String() (out string) {
	out += c.StringName() + "\n"
	out += c.Bytecode.String()
	return out
}

// StringChildren satisfies the requirements of the debug.StringTree interface
// so that related closures can be pretty-printed
func (c *ClosureTemplate) StringChildren() (children []debug.StringTree) {
	for _, enclosed := range c.Enclosed {
		children = append(children, enclosed)
	}
	return children
}

func (c *ClosureTemplate) isObject() {}

var uniqueClosureID int

// MakeEmptyClosureTemplate is a helper function to create a closure template
// from a given set of parameter RegisterTemplates and a Bytecode blob. This
// function should be called during codegen exactly once for each function
// literal
func MakeEmptyClosureTemplate(params []*RegisterTemplate) *ClosureTemplate {
	nextID := uniqueClosureID
	uniqueClosureID++
	template := &ClosureTemplate{
		ID:         nextID,
		Parameters: params,
		Bytecode:   &Bytecode{},
	}
	template.Bytecode.Closure = template
	return template
}

// Closure is an object created during runtime from a ClosureTemplate that binds
// a runtime environment to a Bytecode blob so that any evaluation of that
// Bytecode is done in the context of the bound environment. Closures are
// objects that can be passed, referenced, and manipulated the same as any other
// object in the virtual machine.
type Closure struct {
	Env        *Env
	Parameters []*RegisterTemplate
	Bytecode   *Bytecode
	Template   *ClosureTemplate
}

func (c *Closure) String() string { return "<closure>" }
func (c *Closure) isObject()      {}
