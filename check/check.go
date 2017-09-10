package check

import (
	"fmt"
	"plaid/parser"
	"sort"
)

// A collection of types native to the execution environment
var (
	BuiltinInt Type = TypeIdent{"Int"}
	BuiltinStr      = TypeIdent{"Str"}
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
func checkBinaryExpr(scope *Scope, expr parser.BinaryExpr) Type {
	switch expr.Oper {
	case "+":
		return checkAddition(scope, expr.Left, expr.Right)
	default:
		scope.addError(fmt.Errorf("unknown infix operator '%s'", expr.Oper))
		return TypeError{}
	}
}

func checkAddition(scope *Scope, left parser.Expr, right parser.Expr) Type {
	leftType := checkExpr(scope, left)
	rightType := checkExpr(scope, right)

	if leftType.IsError() || rightType.IsError() {
		return TypeError{}
	}

	typ := BuiltinInt

	if leftType.Equals(BuiltinInt) == false {
		scope.addError(fmt.Errorf("left side must have type %s, got %s", BuiltinInt, leftType))
		typ = TypeError{}
	}

	if rightType.Equals(BuiltinInt) == false {
		scope.addError(fmt.Errorf("right side must have type %s, got %s", BuiltinInt, rightType))
		typ = TypeError{}
	}

	return typ
}

func checkIdentExpr(scope *Scope, expr parser.IdentExpr) Type {
	if scope.hasVariable(expr.Name) {
		return scope.getVariable(expr.Name)
	}

	scope.addError(fmt.Errorf("variable '%s' was used before it was declared", expr.Name))
	return TypeError{}
}

func checkNumberExpr(scope *Scope, expr parser.NumberExpr) Type {
	return BuiltinInt
}

func checkStringExpr(scope *Scope, expr parser.StringExpr) Type {
	return BuiltinStr
}
