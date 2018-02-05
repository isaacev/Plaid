package lang

import (
	"plaid/lang/types"
)

type Scope struct {
	Module   *ModuleVirtual
	Parent   *Scope
	Children map[ASTNode]*Scope
	Local    map[string]types.Type
	Self     types.Function
	Errors   []error
}

func makeXScope(parent *Scope) *Scope {
	scope := &Scope{
		Parent:   parent,
		Children: make(map[ASTNode]*Scope),
		Local:    make(map[string]types.Type),
	}

	if parent != nil {
		scope.Module = parent.Module
	}

	return scope
}

func (s *Scope) HasLocal(name string) bool {
	if _, ok := s.Local[name]; ok {
		return true
	}
	return false
}

func (s *Scope) AddLocal(name string, typ types.Type) {
	s.Local[name] = typ
}

func (s *Scope) Lookup(name string) types.Type {
	if s.HasLocal(name) {
		return s.Local[name]
	}

	if s.Parent != nil {
		return s.Parent.Lookup(name)
	} else if s.Module != nil {
		for _, dep := range s.Module.dependencies {
			if dep.alias == name {
				return dep.module.Exports()
			}
		}
	}

	return nil
}

func (s *Scope) AllErrors() []error {
	errs := s.Errors
	for _, scope := range s.Children {
		errs = append(errs, scope.AllErrors()...)
	}
	return errs
}
