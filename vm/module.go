package vm

import "plaid/types"

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
