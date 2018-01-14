package lang

import (
	"fmt"
	"plaid/lang/printing"
	"plaid/lang/types"
	"sort"
)

type UniqueSymbol struct{}

// Scope describes common methods that any type of scope must implement
type Scope interface {
	printing.StringerTree
	HasParent() bool
	GetParent() Scope
	addChild(Scope) Scope
	HasSelfReference() bool
	GetSelfReference() types.Function
	HasErrors() bool
	GetErrors() []error
	newError(error)
	HasLocalVariable(string) bool
	GetLocalVariableType(string) types.Type
	GetLocalVariableReference(string) *UniqueSymbol
	GetLocalVariableNames() []string
	HasVariable(string) bool
	GetVariableType(string) types.Type
	GetVariableReference(string) *UniqueSymbol
	newVariable(string, types.Type) *UniqueSymbol
}

// GlobalScope exists at the top of the scope tree
type GlobalScope struct {
	imports  []*GlobalScope
	exports  map[string]types.Type
	children []Scope
	errors   []error
	types    map[string]types.Type
	symbols  map[string]*UniqueSymbol
}

// makeGlobalScope is a helper function to quickly build a global scope
func makeGlobalScope() *GlobalScope {
	return &GlobalScope{
		exports: make(map[string]types.Type),
		types:   make(map[string]types.Type),
		symbols: make(map[string]*UniqueSymbol),
	}
}

// addImport exposes another module's exports to the global scope
func (s *GlobalScope) addImport(module *GlobalScope) {
	if module == nil {
		panic("tried to add <nil> as import")
	}

	s.imports = append(s.imports, module)
}

// HasExport returns true if a given name has been mapped to a *Export struct
func (s *GlobalScope) HasExport(name string) bool {
	if _, exists := s.exports[name]; exists {
		return true
	}

	return false
}

// GetExport returns an Export struct if the given variable is exported, returns
// nil otherwise
func (s *GlobalScope) GetExport(name string) types.Type {
	if s.HasExport(name) {
		return s.exports[name]
	}

	return nil
}

// newExport exposes global definitions for use by other modules
func (s *GlobalScope) newExport(name string, typ types.Type) {
	s.exports[name] = typ
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
func (s *GlobalScope) GetSelfReference() types.Function { return types.Function{} }

// HasErrors returns true if any errors have been logged with this scope
func (s *GlobalScope) HasErrors() bool { return len(s.errors) > 0 }

// GetErrors returns any erros that have been logged with this scope
func (s *GlobalScope) GetErrors() []error { return s.errors }

// newError appends another error to the global list of logged errors
func (s *GlobalScope) newError(err error) { s.errors = append(s.errors, err) }

// HasLocalVariable returns true if *this* scope recognizes the given variable
func (s *GlobalScope) HasLocalVariable(name string) bool {
	if _, exists := s.symbols[name]; exists {
		return true
	}

	return false
}

// GetLocalVariableType returns the type of a given variable if it exists in the
// local scope. If the variable cannot be found the method returns nil
func (s *GlobalScope) GetLocalVariableType(name string) types.Type {
	if s.HasLocalVariable(name) {
		return s.types[name]
	}

	return nil
}

// GetLocalVariableReference returns the unique reference identifier of a given
// variable if it exists in the local scope. It the variable cannot be found the
// method returns nil
func (s *GlobalScope) GetLocalVariableReference(name string) *UniqueSymbol {
	if s.HasLocalVariable(name) {
		return s.symbols[name]
	}

	return nil
}

// GetLocalVariableNames returns a list of all locally registered variables
func (s *GlobalScope) GetLocalVariableNames() (names []string) {
	for name := range s.symbols {
		names = append(names, name)
	}
	return names
}

// HasVariable returns true if this or *any parent scope* recognizes the given
// variable. For global scope this is the same as HasLocalvariable
func (s *GlobalScope) HasVariable(name string) bool {
	if s.HasLocalVariable(name) {
		return true
	}

	for _, module := range s.imports {
		if module.HasExport(name) {
			return true
		}
	}

	return false
}

// GetVariableType returns the type associated with the given variable name if
// it can be found in scope. It returns nil if the variable cannot be found in
// the current scope
func (s *GlobalScope) GetVariableType(name string) types.Type {
	if s.HasLocalVariable(name) {
		return s.GetLocalVariableType(name)
	}

	for _, module := range s.imports {
		if module.HasExport(name) {
			return module.GetExport(name)
		}
	}

	return nil
}

// GetVariableReference returns the reigster template associated with the given
// variable if it can be found in scope. It returns nil if the variable cannot
// be found in scope
func (s *GlobalScope) GetVariableReference(name string) *UniqueSymbol {
	if s.HasLocalVariable(name) {
		return s.GetLocalVariableReference(name)
	}

	for _, module := range s.imports {
		if module.HasExport(name) {
			return module.GetVariableReference(name)
		}
	}

	return nil
}

// newVariable registers a new variable with the given name and type and
// generates a unique reference identifier for that variable
func (s *GlobalScope) newVariable(name string, typ types.Type) *UniqueSymbol {
	ref := &UniqueSymbol{}
	s.types[name] = typ
	s.symbols[name] = ref
	return ref
}

func (s *GlobalScope) String() (out string) {
	var padding int
	var names []string
	for name := range s.symbols {
		if len(name) >= padding {
			padding = len(name)

			if s.HasExport(name) {
				padding++
			}
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
		if s.HasExport(name) {
			name = "@" + name
		}
		out += fmt.Sprintf(pattern, name, typ)
	}

	return out
}

// StringerChildren satisfies the StringTree interface
func (s *GlobalScope) StringerChildren() (children []printing.StringerTree) {
	for _, child := range s.children {
		children = append(children, child)
	}
	return children
}

// LocalScope exists inside of function literals
type LocalScope struct {
	parent     Scope
	children   []Scope
	self       types.Function
	types      map[string]types.Type
	references map[string]*UniqueSymbol
}

// makeLocalScope is a helper function to quickly build a local scope that is
// doubly linked to its parent scope
func makeLocalScope(parent Scope, self types.Function) *LocalScope {
	return parent.addChild(&LocalScope{
		parent:     parent,
		self:       self,
		types:      make(map[string]types.Type),
		references: make(map[string]*UniqueSymbol),
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
func (s *LocalScope) GetSelfReference() types.Function { return s.self }

// HasErrors returns true if any errors have been logged with this scope
func (s *LocalScope) HasErrors() bool { return s.parent.HasErrors() }

// GetErrors returns any erros that have been logged with this scope
func (s *LocalScope) GetErrors() []error { return s.parent.GetErrors() }

// newError appends another error to the global list of logged errors
func (s *LocalScope) newError(err error) { s.parent.newError(err) }

// SelfReference returns the type signature of the local function
func (s *LocalScope) SelfReference() types.Type { return nil }

// HasLocalVariable returns true if *this* scope recognizes the given variable
func (s *LocalScope) HasLocalVariable(name string) bool {
	if _, exists := s.references[name]; exists {
		return true
	}

	return false
}

// GetLocalVariableType returns the type associated with the given variable name if
// it can be found in scope. It returns nil if the variable cannot be found in
// the current scope
func (s *LocalScope) GetLocalVariableType(name string) types.Type {
	if s.HasLocalVariable(name) {
		return s.types[name]
	}

	return nil
}

// GetLocalVariableReference returns the unique reference identifier of a given
// variable if it exists in the local scope. It the variable cannot be found the
// method returns nil
func (s *LocalScope) GetLocalVariableReference(name string) *UniqueSymbol {
	if s.HasLocalVariable(name) {
		return s.references[name]
	}

	return nil
}

// GetLocalVariableNames returns a list of all locally registered variables
func (s *LocalScope) GetLocalVariableNames() (names []string) {
	for name := range s.references {
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

// GetVariableType returns the type associated with the given variable name if
// it can be found in scope. It returns nil if the variable cannot be found in
// the current scope
func (s *LocalScope) GetVariableType(name string) types.Type {
	if s.HasLocalVariable(name) {
		return s.GetLocalVariableType(name)
	}

	return s.parent.GetVariableType(name)
}

// GetVariableReference returns the reigster template associated with the given
// variable if it can be found in scope. It returns nil if the variable cannot
// be found in scope
func (s *LocalScope) GetVariableReference(name string) *UniqueSymbol {
	if s.HasLocalVariable(name) {
		return s.GetLocalVariableReference(name)
	}

	return s.parent.GetVariableReference(name)
}

func (s *LocalScope) String() (out string) {
	var padding int
	var names []string
	for name := range s.references {
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

// StringerChildren satisfies the StringTree interface
func (s *LocalScope) StringerChildren() (children []printing.StringerTree) {
	for _, child := range s.children {
		children = append(children, child)
	}
	return children
}

// newVariable registers a new variable with the given name and type and
// generates a unique reference identifier for that variable
func (s *LocalScope) newVariable(name string, typ types.Type) *UniqueSymbol {
	ref := &UniqueSymbol{}
	s.types[name] = typ
	s.references[name] = ref
	return ref
}
