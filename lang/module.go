package lang

import "fmt"

type Module interface {
	Path() string
	Scope() *GlobalScope
	Imports() []Module
	fmt.Stringer
}

type NativeModule struct {
	path  string
	scope *GlobalScope
}

func (m *NativeModule) Path() string        { return m.path }
func (m *NativeModule) Scope() *GlobalScope { return m.scope }
func (m *NativeModule) Imports() []Module   { return nil }

func (m *NativeModule) String() string {
	return m.Path()
}

func BuildNativeModule(name string, exports map[string]Type) *NativeModule {
	return &NativeModule{
		path: name,
		scope: &GlobalScope{
			exports: exports,
		},
	}
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

func (m *VirtualModule) String() string {
	return m.Path()
}
