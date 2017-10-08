package check

import (
	"fmt"
	"plaid/parser"
	"plaid/types"
)

// Scope tracks the symbol table and other data used during the check
type Scope struct {
	parent        *Scope
	errs          []error
	variables     []string
	values        map[string]types.Type
	pendingReturn types.Type
	queue         []struct {
		ret  types.Type
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

func (s *Scope) hasLocalVariable(name string) bool {
	_, exists := s.values[name]
	return exists
}

func (s *Scope) existingVariable(name string) bool {
	if s.hasLocalVariable(name) {
		return true
	} else if s.hasParent() {
		return s.parent.existingVariable(name)
	} else {
		return false
	}
}

func (s *Scope) registerLocalVariable(name string, typ types.Type) {
	s.variables = append(s.variables, name)
	s.values[name] = typ
}

func (s *Scope) getVariable(name string) types.Type {
	if s.hasLocalVariable(name) {
		return s.values[name]
	} else if s.hasParent() {
		return s.parent.getVariable(name)
	} else {
		return nil
	}
}

func (s *Scope) hasPendingReturnType() bool {
	return (s.pendingReturn != nil)
}

func (s *Scope) getPendingReturnType() types.Type {
	return s.pendingReturn
}

func (s *Scope) setPendingReturnType(typ types.Type) {
	s.pendingReturn = typ
}

func (s *Scope) hasBodyQueue() bool {
	return len(s.queue) > 0
}

func (s *Scope) enqueueBody(ret types.Type, expr parser.FunctionExpr) {
	body := struct {
		ret  types.Type
		expr parser.FunctionExpr
	}{ret, expr}
	s.queue = append(s.queue, body)
}

func (s *Scope) dequeueBody() (types.Type, parser.FunctionExpr) {
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

func makeScope(parent *Scope, ret types.Type) *Scope {
	return &Scope{
		parent,
		[]error{},
		[]string{},
		make(map[string]types.Type),
		ret,
		[]struct {
			ret  types.Type
			expr parser.FunctionExpr
		}{},
	}
}
