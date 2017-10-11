package codegen

import (
	"fmt"
	"plaid/types"
	"plaid/vm"
)

// LexicalScope represents a region of the program where all local variables
// have the same lifetime
type LexicalScope struct {
	parent  *LexicalScope
	Local   map[string]*VarRecord
	Global  map[string]*vm.Builtin
	SelfRet types.Type
}

func (ls *LexicalScope) hasParent() bool {
	return (ls.parent != nil)
}

func (ls *LexicalScope) hasGlobalVariable(name string) bool {
	_, exists := ls.Global[name]
	return exists
}

func (ls *LexicalScope) addGlobalVariable(name string, builtin *vm.Builtin) {
	if ls.hasGlobalVariable(name) {
		panic(fmt.Sprintf("variable '%s' redeclared locally", name))
	}

	ls.Global[name] = builtin
}

func (ls *LexicalScope) getGlobalVariable(name string) *vm.Builtin {
	if ls.hasGlobalVariable(name) {
		return ls.Global[name]
	}

	panic(fmt.Sprintf("global variable '%s' is not in scope", name))
}

func (ls *LexicalScope) hasLocalVariable(name string) bool {
	_, exists := ls.Local[name]
	return exists
}

func (ls *LexicalScope) addLocalVariable(name string, typ types.Type) *VarRecord {
	if ls.hasLocalVariable(name) {
		panic(fmt.Sprintf("variable '%s' redeclared locally", name))
	}

	varRecordID++
	id := varRecordID
	record := &VarRecord{id, name, typ}
	ls.Local[name] = record
	return record
}

func (ls *LexicalScope) getVariable(name string) *VarRecord {
	if ls.hasLocalVariable(name) {
		return ls.Local[name]
	} else if ls.hasParent() {
		return ls.parent.getVariable(name)
	} else {
		panic(fmt.Sprintf("variable '%s' is not in scope", name))
	}
}

func makeLexicalScope(parent *LexicalScope) *LexicalScope {
	globals := make(map[string]*vm.Builtin)
	if parent != nil {
		globals = parent.Global
	}

	return &LexicalScope{
		parent,
		make(map[string]*VarRecord),
		globals,
		nil,
	}
}
