package lang

import (
	"fmt"
	"plaid/lang/types"
)

type Object interface {
	fmt.Stringer
	Value() interface{}
	isObject()
}

type ObjectNone struct{}

func (o ObjectNone) Value() interface{} { return nil }
func (o ObjectNone) String() string     { return "<none>" }
func (o ObjectNone) isObject()          {}

type ObjectInt struct {
	val int64
}

func (o ObjectInt) Value() interface{} { return o.val }
func (o ObjectInt) String() string     { return fmt.Sprintf("%d", o.val) }
func (o ObjectInt) isObject()          {}

type ObjectStr struct {
	val string
}

func (o ObjectStr) Value() interface{} { return o.val }
func (o ObjectStr) String() string     { return fmt.Sprintf("\"%s\"", o.val) }
func (o ObjectStr) isObject()          {}

type ObjectBool struct {
	val bool
}

func (o ObjectBool) Value() interface{} { return o.val }
func (o ObjectBool) String() string     { return fmt.Sprintf("%t", o.val) }
func (o ObjectBool) isObject()          {}

type ObjectBuiltin struct {
	typ types.Type
	val func(args []Object) (Object, error)
}

func (o ObjectBuiltin) Type() types.Type   { return o.typ }
func (o ObjectBuiltin) Value() interface{} { return o.val }
func (o ObjectBuiltin) String() string     { return "<builtin>" }
func (o ObjectBuiltin) isObject()          {}

type ObjectFunction struct {
	params   []string
	bytecode Bytecode
}

func (o ObjectFunction) Value() interface{} { return o.bytecode }
func (o ObjectFunction) String() string     { return "<function>" }
func (o ObjectFunction) isObject()          {}

type ObjectClosure struct {
	// Env      *Env
	params   []string
	bytecode *Bytecode
}

func (o ObjectClosure) Value() interface{} { return o.bytecode }
func (o ObjectClosure) String() string     { return "<closure>" }
func (o ObjectClosure) isObject()          {}

type ObjectStruct struct {
	fields map[string]Object
}

func (o ObjectStruct) Value() interface{}        { return nil }
func (o ObjectStruct) String() string            { return "<struct>" }
func (o ObjectStruct) isObject()                 {}
func (o ObjectStruct) Member(name string) Object { return o.fields[name] }
