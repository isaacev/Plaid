package lang

import (
	"fmt"
	"plaid/lang/printing"
	"plaid/lang/types"
)

type Module interface {
	Path() string
	Scope() *GlobalScope
	Imports() []Module
	link(Module)
	fmt.Stringer
}

type NativeModule struct {
	path    string
	scope   *GlobalScope
	library *Library
}

func (m *NativeModule) Path() string        { return m.path }
func (m *NativeModule) Scope() *GlobalScope { return m.scope }
func (m *NativeModule) Imports() []Module   { return nil }
func (m *NativeModule) link(mod Module)     {}

func (m *NativeModule) String() (out string) {
	return m.Path()
}

func MakeNativeModule(name string, types map[string]types.Type, objects map[string]func(args []Object) (Object, error)) *NativeModule {
	mod := &NativeModule{
		path:  name,
		scope: makeGlobalScope(),
	}

	for name, typ := range types {
		if val, ok := objects[name]; ok {
			sym := mod.scope.newExportObject(name, typ, ObjectBuiltin{
				typ: typ,
				val: val,
			})
			fmt.Println("%s -> %p\n", name, sym)
		} else {
			panic(fmt.Sprintf("malformed library"))
		}
	}

	return mod
}

type VirtualModule struct {
	path    string
	ast     *RootNode
	scope   *GlobalScope
	imports []Module
}

func (m *VirtualModule) Path() string        { return m.path }
func (m *VirtualModule) Scope() *GlobalScope { return m.scope }
func (m *VirtualModule) Imports() []Module   { return m.imports }
func (m *VirtualModule) link(mod Module)     { m.imports = append(m.imports, mod) }

func (m *VirtualModule) String() (out string) {
	out += fmt.Sprintf("path: %s\n", m.Path())
	for _, imp := range m.Imports() {
		out += fmt.Sprintf("uses: %s\n", imp.Path())
	}
	out += printing.TreeToString(m.scope)
	return out
}
