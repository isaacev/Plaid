package linker

import "testing"

func TestAddImport(t *testing.T) {
	mod1 := &Module{}
	mod2 := &Module{}

	mod1.AddImport(mod2)

	if mod1.Imports[0] != mod2 {
		t.Errorf("Expected to add 'mod2' to Module.Imports")
	}
}

func TestModuleString(t *testing.T) {
	mod := &Module{Name: "foo"}
	expectString(t, mod.String(), "<module foo>")
}

func expectString(t *testing.T, got string, exp string) {
	if exp != got {
		t.Errorf("Expected '%s', got '%s'", exp, got)
	}
}

func expectBool(t *testing.T, got bool, exp bool) {
	if exp != got {
		t.Errorf("Expected %t, got %t", exp, got)
	}
}
