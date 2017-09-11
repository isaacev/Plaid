package check

import (
	"fmt"
	"plaid/parser"
)

// Scope tracks the symbol table and other data used during the check
type Scope struct {
	parent        *Scope
	errs          []error
	variables     []string
	values        map[string]Type
	pendingReturn Type
	queue         []struct {
		ret  Type
		expr parser.FunctionExpr
	}
}

func (s *Scope) hasParent() bool {
	return (s.parent != nil)
}

// Errors returns a list of errors detected during the check
func (s *Scope) Errors() []error {
	if s.hasParent() {
		return s.parent.Errors()
	}

	return s.errs
}

func (s *Scope) addError(err error) {
	if s.hasParent() {
		s.parent.addError(err)
	} else {
		s.errs = append(s.errs, err)
	}
}

func (s *Scope) hasVariable(name string) bool {
	_, exists := s.values[name]

	if exists == false && s.hasParent() {
		return s.parent.hasVariable(name)
	}

	return exists
}

func (s *Scope) registerLocalVariable(name string, typ Type) {
	s.variables = append(s.variables, name)
	s.values[name] = typ
}

func (s *Scope) getVariable(name string) Type {
	return s.values[name]
}

func (s *Scope) hasPendingReturnType() bool {
	return (s.pendingReturn != nil)
}

func (s *Scope) getPendingReturnType() Type {
	return s.pendingReturn
}

func (s *Scope) setPendingReturnType(typ Type) {
	s.pendingReturn = typ
}

func (s *Scope) hasBodyQueue() bool {
	return len(s.queue) > 0
}

func (s *Scope) enqueueBody(ret Type, expr parser.FunctionExpr) {
	body := struct {
		ret  Type
		expr parser.FunctionExpr
	}{ret, expr}
	s.queue = append(s.queue, body)
}

func (s *Scope) dequeueBody() (Type, parser.FunctionExpr) {
	body := s.queue[0]
	s.queue = s.queue[1:]
	return body.ret, body.expr
}

func (s *Scope) String() string {
	var out string
	for _, name := range s.variables {
		out += fmt.Sprintf("%s : %s\n", name, s.values[name])
	}
	return out
}

func makeScope(parent *Scope, ret Type) *Scope {
	return &Scope{
		parent,
		[]error{},
		[]string{},
		make(map[string]Type),
		ret,
		[]struct {
			ret  Type
			expr parser.FunctionExpr
		}{},
	}
}
