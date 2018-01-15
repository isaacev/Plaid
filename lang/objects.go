package lang

import (
	"fmt"
	"go/types"
)

type Object interface {
	fmt.Stringer
	isObject()
}

type ObjectNone struct{}

func (o ObjectNone) String() string { return "<none>" }
func (o ObjectNone) isObject()      {}

type ObjectInt struct {
	Val int64
}

func (o ObjectInt) String() string { return fmt.Sprintf("%d", o.Val) }
func (o ObjectInt) isObject()      {}

type ObjectStr struct {
	Val string
}

func (o ObjectStr) String() string { return o.Val }
func (o ObjectStr) isObject()      {}

type ObjectBool struct {
	Val bool
}

func (o ObjectBool) String() string { return fmt.Sprintf("%t", o.Val) }
func (o ObjectBool) isObject()      {}

type ObjectBuiltin struct {
	Type types.Type
	Val  func(args []Object) (Object, error)
}

func (o ObjectBuiltin) String() string { return "<builtin>" }
func (o ObjectBuiltin) isObject()      {}

type ObjectFunction struct {
	// Params   []*UniqueSymbol
	// Bytecode *Bytecode
}

func (o ObjectFunction) String() string { return "<function>" }
func (o ObjectFunction) isObject()      {}

type ObjectClosure struct {
	// Env      *Env
	// Params   []*UniqueSymbol
	// Bytecode *Bytecode
}

func (o ObjectClosure) String() string { return "<closure>" }
func (o ObjectClosure) isObject()      {}
