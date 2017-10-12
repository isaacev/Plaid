package scope

import (
	"fmt"
	"plaid/debug"
	"plaid/types"
	"plaid/vm"
	"sort"
)

// Scope describes common methods that any type of scope must implement
type Scope interface {
	debug.StringTree
	HasParent() bool
	GetParent() Scope
	addChild(Scope) Scope
	HasSelfReference() bool
	GetSelfReference() types.TypeFunction
	HasErrors() bool
	GetErrors() []error
	NewError(error)
	HasLocalVariable(string) bool
	GetLocalVariable(string) (*vm.RegisterTemplate, types.Type, error)
	GetLocalVariableNames() []string
	HasVariable(string) bool
	GetVariable(string) (*vm.RegisterTemplate, types.Type, error)
	GetVariableType(string) types.Type
	NewVariable(string, types.Type) *vm.RegisterTemplate
}

// GlobalScope exists at the top of the scope tree
type GlobalScope struct {
	children  []Scope
	errors    []error
	types     map[string]types.Type
	registers map[string]*vm.RegisterTemplate
}

// MakeGlobalScope is a helper function to quickly build a global scope
func MakeGlobalScope() *GlobalScope {
	return &GlobalScope{
		types:     make(map[string]types.Type),
		registers: make(map[string]*vm.RegisterTemplate),
	}
}

// HasParent returns true if the current scope has a parent scope
func (s *GlobalScope) HasParent() bool { return false }

// GetParent returns the parent scope of the current scope
func (s *GlobalScope) GetParent() Scope { return nil }

func (s *GlobalScope) addChild(child Scope) Scope {
	s.children = append(s.children, child)
	return child
}

// HasSelfReference returns false because no function can contain a global scope
func (s *GlobalScope) HasSelfReference() bool { return false }

// GetSelfReference returns an empty function type because no function can
// contain a global scope
func (s *GlobalScope) GetSelfReference() types.TypeFunction { return types.TypeFunction{} }

// HasErrors returns true if any errors have been logged with this scope
func (s *GlobalScope) HasErrors() bool { return len(s.errors) > 0 }

// GetErrors returns any erros that have been logged with this scope
func (s *GlobalScope) GetErrors() []error { return s.errors }

// NewError appends another error to the global list of logged errors
func (s *GlobalScope) NewError(err error) { s.errors = append(s.errors, err) }

// HasLocalVariable returns true if *this* scope recognizes the given variable
func (s *GlobalScope) HasLocalVariable(name string) bool {
	if _, exists := s.registers[name]; exists {
		return true
	}

	return false
}

// GetLocalVariable returns the register template associated with the given
// variable or an error if the given variable is not local
func (s *GlobalScope) GetLocalVariable(name string) (*vm.RegisterTemplate, types.Type, error) {
	if s.HasLocalVariable(name) {
		return s.registers[name], s.types[name], nil
	}

	return nil, nil, fmt.Errorf("variable '%s' is not in scope", name)
}

// GetLocalVariableNames returns a list of all locally registered variables
func (s *GlobalScope) GetLocalVariableNames() (names []string) {
	for name := range s.registers {
		names = append(names, name)
	}
	return names
}

// HasVariable returns true if this or *any parent scope* recognizes the given
// variable. For global scope this is the same as HasLocalvariable
func (s *GlobalScope) HasVariable(name string) bool {
	return s.HasLocalVariable(name)
}

// GetVariable returns the register template associated with the given variable
// or an error if the given variable is not recognized by this or any of its
// parent scopes. For global scope this is the same as GetLocalVariable
func (s *GlobalScope) GetVariable(name string) (*vm.RegisterTemplate, types.Type, error) {
	return s.GetLocalVariable(name)
}

// GetVariableType returns the associated type of a variable if its in scope or
// returns the error type if it's not in scope
func (s *GlobalScope) GetVariableType(name string) types.Type {
	if s.HasVariable(name) {
		return s.types[name]
	}

	return types.TypeError{}
}

// NewVariable registers a new variable with the given name and type and
// generates a unique register template for that variable
func (s *GlobalScope) NewVariable(name string, typ types.Type) *vm.RegisterTemplate {
	reg := vm.MakeRegisterTemplate(name)
	s.types[name] = typ
	s.registers[name] = reg
	return reg
}

func (s *GlobalScope) String() (out string) {
	var padding int
	var names []string
	for name := range s.registers {
		if len(name) > padding {
			padding = len(name)
		}
		names = append(names, name)
	}

	sort.Strings(names)
	pattern := fmt.Sprintf("%%-%ds : %%s", padding)
	for i, name := range names {
		if i > 0 {
			out += "\n"
		}
		typ := s.types[name]
		out += fmt.Sprintf(pattern, name, typ)
	}

	return out
}

// StringChildren satisfies the debug.StringTree interface
func (s *GlobalScope) StringChildren() (children []debug.StringTree) {
	for _, child := range s.children {
		children = append(children, child)
	}
	return children
}

// LocalScope exists inside of function literals
type LocalScope struct {
	parent    Scope
	children  []Scope
	self      types.TypeFunction
	types     map[string]types.Type
	registers map[string]*vm.RegisterTemplate
}

// MakeLocalScope is a helper function to quickly build a local scope that is
// doubly linked to its parent scope
func MakeLocalScope(parent Scope, self types.TypeFunction) *LocalScope {
	return parent.addChild(&LocalScope{
		parent:    parent,
		self:      self,
		types:     make(map[string]types.Type),
		registers: make(map[string]*vm.RegisterTemplate),
	}).(*LocalScope)
}

// HasParent returns true if the current scope has a parent scope
func (s *LocalScope) HasParent() bool { return s.parent != nil }

// GetParent returns the parent scope of the current scope
func (s *LocalScope) GetParent() Scope { return s.parent }

func (s *LocalScope) addChild(child Scope) Scope {
	s.children = append(s.children, child)
	return child
}

// HasSelfReference returns true because any local scope is inside a function
func (s *LocalScope) HasSelfReference() bool { return true }

// GetSelfReference returns the type signature of the current function
func (s *LocalScope) GetSelfReference() types.TypeFunction { return s.self }

// HasErrors returns true if any errors have been logged with this scope
func (s *LocalScope) HasErrors() bool { return s.parent.HasErrors() }

// GetErrors returns any erros that have been logged with this scope
func (s *LocalScope) GetErrors() []error { return s.parent.GetErrors() }

// NewError appends another error to the global list of logged errors
func (s *LocalScope) NewError(err error) { s.parent.NewError(err) }

// SelfReference returns the type signature of the local function
func (s *LocalScope) SelfReference() types.Type { return nil }

// HasLocalVariable returns true if *this* scope recognizes the given variable
func (s *LocalScope) HasLocalVariable(name string) bool {
	if _, exists := s.registers[name]; exists {
		return true
	}

	return false
}

// GetLocalVariable returns the register template associated with the given
// variable or an error if the given variable is not local
func (s *LocalScope) GetLocalVariable(name string) (*vm.RegisterTemplate, types.Type, error) {
	if s.HasLocalVariable(name) {
		return s.registers[name], s.types[name], nil
	}

	return nil, nil, fmt.Errorf("variable '%s' is not in scope", name)
}

// GetLocalVariableNames returns a list of all locally registered variables
func (s *LocalScope) GetLocalVariableNames() (names []string) {
	for name := range s.registers {
		names = append(names, name)
	}
	return names
}

// HasVariable returns true if this or *any parent scope* recognizes the given
// variable
func (s *LocalScope) HasVariable(name string) bool {
	if s.HasLocalVariable(name) {
		return true
	}

	return s.parent.HasVariable(name)
}

// GetVariable returns the register template associated with the given variable
// or an error if the given variable is not recognized by this or any of its
// parent scopes
func (s *LocalScope) GetVariable(name string) (*vm.RegisterTemplate, types.Type, error) {
	if s.HasLocalVariable(name) {
		return s.GetLocalVariable(name)
	}

	return s.parent.GetVariable(name)
}

// GetVariableType returns the associated type of a variable if its in scope or
// returns the error type if it's not in scope
func (s *LocalScope) GetVariableType(name string) types.Type {
	if s.HasVariable(name) {
		_, typ, _ := s.GetVariable(name)
		return typ
	}

	return types.TypeError{}
}

func (s *LocalScope) String() (out string) {
	var padding int
	var names []string
	for name := range s.registers {
		if len(name) > padding {
			padding = len(name)
		}
		names = append(names, name)
	}

	sort.Strings(names)
	pattern := fmt.Sprintf("%%-%ds : %%s", padding)
	for i, name := range names {
		if i > 0 {
			out += "\n"
		}
		typ := s.types[name]
		out += fmt.Sprintf(pattern, name, typ)
	}

	return out
}

// StringChildren satisfies the debug.StringTree interface
func (s *LocalScope) StringChildren() (children []debug.StringTree) {
	for _, child := range s.children {
		children = append(children, child)
	}
	return children
}

// NewVariable registers a new variable with the given name and type and
// generates a unique register template for that variable
func (s *LocalScope) NewVariable(name string, typ types.Type) *vm.RegisterTemplate {
	reg := vm.MakeRegisterTemplate(name)
	s.types[name] = typ
	s.registers[name] = reg
	return reg
}
