package linker

import (
	"fmt"
	"plaid/parser"
	"plaid/scope"
	"plaid/vm"
)

// Module relates a program with its dependencies and its syntactic structure
type Module struct {
	Name    string
	AST     *parser.Program
	Scope   *scope.GlobalScope
	Imports []*Module
}

// AddImport does some stuff
func (m *Module) AddImport(mod *Module) {
	m.Imports = append(m.Imports, mod)
	m.Scope.AddImport(mod.Scope)
}

func (m *Module) String() string {
	return fmt.Sprintf("<module %s>", m.Name)
}

// ConvertModule transforms a Module used by the VM into a Module used by the
// type system including linked imports
func ConvertModule(mod *vm.Module) *Module {
	global := scope.MakeGlobalScope()
	for name, exp := range mod.Exports {
		global.Export(name, exp.Type)
	}

	var deps []*Module
	for _, m := range mod.Imports {
		if m == mod {
			panic("not a DAG")
		}

		deps = append(deps, ConvertModule(m))
	}

	return &Module{
		Name:    mod.Name,
		Scope:   global,
		Imports: deps,
	}
}
