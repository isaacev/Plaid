package vm

import (
	"fmt"
)

// Object represents all objects that can exist in the machine
type Object interface {
	String() string
	isObject()
}

// ObjectNone is used for builtin operations that require a value to push onto
// the stack
type ObjectNone struct{}

func (on *ObjectNone) String() string { return "<none>" }
func (on *ObjectNone) isObject()      {}

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

// ObjectInt represents any integer object
type ObjectInt struct {
	Val int64
}

func (oi *ObjectInt) String() string { return fmt.Sprintf("%d", oi.Val) }
func (oi *ObjectInt) isObject()      {}

// ObjectStr represents any string object
type ObjectStr struct {
	Val string
}

func (os *ObjectStr) String() string { return os.Val }
func (os *ObjectStr) isObject()      {}

// ObjectBool represents any boolean object
type ObjectBool struct {
	Val bool
}

func (ob *ObjectBool) String() string { return fmt.Sprintf("%t", ob.Val) }
func (ob *ObjectBool) isObject()      {}

// ObjectBuiltin wraps a builtin function as an object that can be manipulated
// the same as a native object
type ObjectBuiltin struct {
	Val *Builtin
}

func (ob *ObjectBuiltin) String() string { return "<builtin>" }
func (ob *ObjectBuiltin) isObject()      {}
