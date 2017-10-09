package vm

import "plaid/types"

type builtinFunc func(args []Object) (Object, error)

// Builtin bundles a type and function
type Builtin struct {
	Type types.Type
	Func builtinFunc
}

// Library is a bundle of builtin functions mapped to the respective names
type Library map[string]*Builtin
