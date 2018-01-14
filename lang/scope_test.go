package lang

import (
	"fmt"
	"plaid/lang/types"
	"testing"
)

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

func TestScopeNewVariable(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("foo", types.Ident{Name: "Bar"})
	if scope.HasVariable("foo") {
		expectEquivalentType(t, scope.GetVariableType("foo"), types.Ident{Name: "Bar"})
	} else {
		t.Errorf("Expected key '%s' in Scope#variables, none found", "foo")
	}
}

func TestScopeGetVariable(t *testing.T) {
	parent := makeGlobalScope()
	child := makeLocalScope(parent, types.Function{})
	parent.newVariable("foo", types.Ident{Name: "Bar"})
	expectEquivalentType(t, child.GetVariableType("foo"), types.Ident{Name: "Bar"})
}

func TestScopeString(t *testing.T) {
	scope := makeGlobalScope()
	scope.newVariable("num", types.Ident{Name: "Int"})
	scope.newVariable("test", types.Ident{Name: "Bool"})
	scope.newVariable("coord", types.Tuple{Children: []types.Type{types.Ident{Name: "Int"}, types.Ident{Name: "Int"}}})

	expectString(t, scope.String(), "coord : (Int Int)\nnum   : Int\ntest  : Bool")
}

func expectNoScopeErrors(t *testing.T, scope Scope) {
	if scope.HasErrors() {
		for i, err := range scope.GetErrors() {
			t.Errorf("%d '%s'", i, err)
		}

		t.Fatalf("Expected no errors, found %d", len(scope.GetErrors()))
	}
}
