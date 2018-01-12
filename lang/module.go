package lang

type Module interface {
	String() string
	Scope() *GlobalScope
	Imports() []Module
}

type NativeModule struct {
	name  string
	scope *GlobalScope
}

func (m *NativeModule) String() string      { return m.name }
func (m *NativeModule) Scope() *GlobalScope { return m.scope }
func (m *NativeModule) Imports() []Module   { return nil }

func BuildNativeModule(name string, exports map[string]Type) *NativeModule {
	return &NativeModule{
		name: name,
		scope: &GlobalScope{
			exports: exports,
		},
	}
}

type VirtualModule struct {
	name    string
	ast     *RootNode
	scope   *GlobalScope
	imports []Module
}

func (m *VirtualModule) String() string      { return m.name }
func (m *VirtualModule) Scope() *GlobalScope { return m.scope }
func (m *VirtualModule) Imports() []Module   { return m.imports }
