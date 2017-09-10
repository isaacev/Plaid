package check

import (
	"fmt"
	"plaid/parser"
	"sort"
)

)

// Check takes an existing abstract syntax tree and performs type checks and
// other correctness checks. It returns a list of any errors that were
// discovered inside the AST
func Check(prog parser.Program) *Scope {
	global := makeScope(nil)
	checkProgram(global, prog)
	return global
}

// Scope tracks the symbol table and other data used during the check
type Scope struct {
	parent    *Scope
	variables map[string]Type
	Errs      []error
}

func (s *Scope) hasParent() bool {
	return (s.parent != nil)
}

func (s *Scope) registerVariable(name string, typ Type) {
	s.variables[name] = typ
}

func (s *Scope) hasVariable(name string) bool {
	_, exists := s.variables[name]
	return exists
}

func (s *Scope) getVariable(name string) Type {
	return s.variables[name]
}

func (s *Scope) addError(err error) {
	if s.hasParent() {
		s.parent.addError(err)
	} else {
		s.Errs = append(s.Errs, err)
	}
}

func (s *Scope) String() string {
	names := []string{}
	for name := range s.variables {
		names = append(names, name)
	}
	sort.Strings(names)

	out := "+----------+--------------+\n"
	out += "| Var      | Type         |\n"
	out += "| -------- | ------------ |\n"
	for _, name := range names {
		out += fmt.Sprintf("| %-8s | %-12s |\n", name, s.variables[name])
	}
	out += "+----------+--------------+\n"
	return out
}

func makeScope(parent *Scope) *Scope {
	scope := &Scope{
		parent,
		make(map[string]Type),
		[]error{},
	}

	return scope
}

func checkProgram(scope *Scope, prog parser.Program) {
	for _, stmt := range prog.Stmts {
		checkStmt(scope, stmt)
	}
}

func checkStmt(scope *Scope, stmt parser.Stmt) {
	switch stmt := stmt.(type) {
	}
}

}
