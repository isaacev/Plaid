package vm

import (
	"fmt"
)

// Object represents all objects that can exist in the machine
type Object interface {
	String() string
	isObject()
}

// ClosureTemplate is generated in the codegen stage and encapsulates the scope
// and bytecode data required to build a closure during execution
type ClosureTemplate struct {
	Parameters []*CellTemplate
	Bytecode   *Bytecode
}

func (ct *ClosureTemplate) String() string { return "<closure template>" }
func (ct *ClosureTemplate) isObject()      {}

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
