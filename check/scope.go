package check

import (
	"fmt"
	"plaid/types"
)

// Scope tracks the symbol table and other data used during the check
type Scope struct {
	parent    *Scope
	children  []*Scope
	errs      []error
	variables []string
	values    map[string]types.Type
	self      *types.TypeFunction
}

func (s *Scope) hasParent() bool {
	return (s.parent != nil)
}

func (s *Scope) extend(self *types.TypeFunction) *Scope {
	child := makeScope(s, self)
	s.children = append(s.children, child)
	return child
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

func (s *Scope) String() string {
	var out string
	for i, name := range s.variables {
		if i > 0 {
			out += "\n"
		}

		out += fmt.Sprintf("%s : %s", name, s.values[name])
	}
	return out
}

func makeScope(parent *Scope, self *types.TypeFunction) *Scope {
	return &Scope{
		parent,
		[]*Scope{},
		[]error{},
		[]string{},
		make(map[string]types.Type),
		self,
	}
}
