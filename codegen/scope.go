package codegen

import (
	"fmt"
	"plaid/check"
)

// LexicalScope represents a region of the program where all local variables
// have the same lifetime
type LexicalScope struct {
	parent *LexicalScope
	Local  map[string]*VarRecord
}

func (ls *LexicalScope) hasParent() bool {
	return (ls.parent != nil)
}

func (ls *LexicalScope) hasLocalVariable(name string) bool {
	_, exists := ls.Local[name]
	return exists
}

func (ls *LexicalScope) addLocalVariable(name string, typ check.Type) *VarRecord {
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
	return &LexicalScope{
		parent,
		make(map[string]*VarRecord),
	}
}
