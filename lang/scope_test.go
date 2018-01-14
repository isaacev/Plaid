package lang

import (
	"fmt"
	"plaid/lang/printing"
	"plaid/lang/types"
	"testing"
)

func TestScopeAddImport(t *testing.T) {
	root := makeGlobalScope()
	imp := makeGlobalScope()
	root.addImport(imp)

	expectSame(t, root.imports[0], imp)

	exp := "tried to add <nil> as import"
	defer func() {
		if got := recover(); got == nil {
			t.Errorf("Expected failure when importing <nil>")
		} else if got != exp {
			t.Errorf("Expected panic '%s', got '%s'", exp, got)
		}
	}()
	root.addImport(nil)
}

func TestScopeHasExport(t *testing.T) {
	root := makeGlobalScope()
	root.newExport("foo", types.TypeNativeBool)

	expectBool(t, root.HasExport("foo"), true)
	expectBool(t, root.HasExport("bar"), false)
}

func TestScopeGetExport(t *testing.T) {
	root := makeGlobalScope()
	root.newExport("foo", types.TypeNativeBool)

	expectSame(t, root.GetExport("foo"), types.TypeNativeBool)
	expectSame(t, root.GetExport("bar"), nil)
}

func TestScopeHasParent(t *testing.T) {
	root := makeGlobalScope()
	child := makeLocalScope(root, types.Function{})
	expectBool(t, root.HasParent(), false)
	expectBool(t, child.HasParent(), true)
}

func TestScopeErrors(t *testing.T) {
	root := makeGlobalScope()
	child := makeLocalScope(root, types.Function{})
	child.newError(fmt.Errorf("foo bar baz"))
	expectNthError(t, root, 0, "foo bar baz")
	expectNthError(t, child, 0, "foo bar baz")
}

func TestScopeAddError(t *testing.T) {
	scope := makeGlobalScope()
	expectNoScopeErrors(t, scope)
	scope.newError(fmt.Errorf("a semantic analysis error"))
	expectNthError(t, scope, 0, "a semantic analysis error")
}

func TestScopeHasLocalVariable(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.Ident{Name: "Bar"})
	expectBool(t, scope.HasLocalVariable("foo"), true)
	expectBool(t, scope.HasLocalVariable("baz"), false)
}

func TestScopeGetLocalVariableType(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.TypeNativeBool)
	expectEquivalentType(t, scope.GetLocalVariableType("foo"), types.TypeNativeBool)
	expectSame(t, scope.GetLocalVariableType("bar"), nil)
}

func TestScopeGetLocalVariableReference(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.TypeNativeBool)
	expectSame(t, scope.GetLocalVariableReference("foo"), scope.symbols["foo"])
	expectBool(t, scope.GetLocalVariableReference("bar") == nil, true)
}

func TestScopeGetLocalVariableNames(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.TypeNativeBool)
	scope.newVariable("bar", types.TypeNativeInt)
	scope.newVariable("baz", types.TypeNativeStr)

	names := scope.GetLocalVariableNames()
	expectSame(t, len(names), 3)
}

func TestScopeHasVariable(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.TypeNativeBool)
	expectBool(t, scope.HasVariable("foo"), true)
	expectBool(t, scope.HasVariable("bar"), false)

	imp := makeGlobalScope()
	scope.addImport(imp)
	imp.newExport("bar", types.TypeNativeInt)
	expectBool(t, scope.HasVariable("bar"), true)
	expectBool(t, scope.HasVariable("baz"), false)
}

func TestScopeGetVariableType(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.TypeNativeBool)
	expectSame(t, scope.GetVariableType("foo"), types.TypeNativeBool)
	expectBool(t, scope.GetVariableType("bar") == nil, true)

	imp := makeGlobalScope()
	scope.addImport(imp)
	imp.newExport("bar", types.TypeNativeInt)
	expectSame(t, scope.GetVariableType("bar"), types.TypeNativeInt)
	expectBool(t, scope.GetVariableType("baz") == nil, true)
}

func TestScopeGetVariableReference(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.TypeNativeBool)
	expectSame(t, scope.GetVariableReference("foo"), scope.symbols["foo"])

	imp := makeGlobalScope()
	scope.addImport(imp)
	imp.newExport("bar", types.TypeNativeInt)
	expectSame(t, scope.GetVariableReference("bar"), imp.symbols["bar"])
	expectBool(t, scope.GetVariableReference("baz") == nil, true)
}

func TestScopeNewVariable(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.Ident{Name: "Bar"})
	if scope.HasVariable("foo") {
		expectEquivalentType(t, scope.GetVariableType("foo"), types.Ident{Name: "Bar"})
	} else {
		t.Errorf("Expected key '%s' in Scope#variables, none found", "foo")
	}
}

func TestScopeString(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("num", types.Ident{Name: "Int"})
	scope.newVariable("test", types.Ident{Name: "Bool"})
	scope.newVariable("coord", types.Tuple{Children: []types.Type{types.Ident{Name: "Int"}, types.Ident{Name: "Int"}}})
	scope.newExport("num", types.TypeNativeBool)
	expectString(t, scope.String(), "coord : (Int Int)\n@num  : Int\ntest  : Bool")
}

func TestScopeStringerChildren(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.TypeNativeInt)
	makeLocalScope(scope, types.Function{})
	expectString(t, printing.TreeToString(scope), "╭─\n┤ foo : Int\n│ ╭─\n╰─┤ \n  ╰─")
}

func expectNoScopeErrors(t *testing.T, scope Scope) {
	if scope.HasErrors() {
		for i, err := range scope.GetErrors() {
			t.Errorf("%d '%s'", i, err)
		}

		t.Fatalf("Expected no errors, found %d", len(scope.GetErrors()))
	}
}
