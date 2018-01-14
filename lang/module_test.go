package lang

import "testing"

func TestNativeModule(t *testing.T) {
	mod := &NativeModule{
		path:  "foo/bar",
		scope: &GlobalScope{},
	}

	expectSame(t, mod.Path(), mod.path)
	expectSame(t, mod.Scope(), mod.scope)
	expectSame(t, len(mod.Imports()), 0)
	expectSame(t, mod.String(), mod.path)
}

func TestVirtualModule(t *testing.T) {
	mod := &VirtualModule{
		path:  "foo/bar",
		ast:   &RootNode{},
		scope: &GlobalScope{},
		imports: []Module{
			&NativeModule{path: "zip/zap"},
		},
	}

	expectSame(t, mod.Path(), mod.path)
	expectSame(t, mod.Scope(), mod.scope)
	expectSame(t, mod.Imports()[0], mod.imports[0])
	expectSame(t, mod.String(), "path: foo/bar\nuses: zip/zap\n╭─\n┤ \n╰─")
}
