package check

import (
	"fmt"
	"plaid/types"
)

// Scope tracks the symbol table and other data used during the check
type Scope struct {
	Parent    *Scope
	Children  []*Scope
	errs      []error
	Variables []string
	Values    map[string]types.Type
	Self      *types.TypeFunction
}

func (s *Scope) hasParent() bool {
	return (s.Parent != nil)
}

func (s *Scope) extend(self *types.TypeFunction) *Scope {
	child := makeScope(s, self)
	s.Children = append(s.Children, child)
	return child
}

// Errors returns a list of errors detected during the check
func (s *Scope) Errors() []error {
	if s.hasParent() {
		return s.Parent.Errors()
	}

	return s.errs
}

func (s *Scope) addError(err error) {
	if s.hasParent() {
		s.Parent.addError(err)
	} else {
		s.errs = append(s.errs, err)
	}
}

func (s *Scope) hasLocalVariable(name string) bool {
	_, exists := s.Values[name]
	return exists
}

func (s *Scope) existingVariable(name string) bool {
	if s.hasLocalVariable(name) {
		return true
	} else if s.hasParent() {
		return s.Parent.existingVariable(name)
	} else {
		return false
	}
}

func (s *Scope) registerLocalVariable(name string, typ types.Type) {
	s.Variables = append(s.Variables, name)
	s.Values[name] = typ
}

func (s *Scope) getVariable(name string) types.Type {
	if s.hasLocalVariable(name) {
		return s.Values[name]
	} else if s.hasParent() {
		return s.Parent.getVariable(name)
	} else {
		return nil
	}
}

func (s *Scope) String() string {
	var out string
	for i, name := range s.Variables {
		if i > 0 {
			out += "\n"
		}

		out += fmt.Sprintf("%s : %s", name, s.Values[name])
	}
	return out
}

func makeScope(parent *Scope, self *types.TypeFunction) *Scope {
	return &Scope{
		Parent: parent,
		Values: make(map[string]types.Type),
		Self:   self,
	}
}
