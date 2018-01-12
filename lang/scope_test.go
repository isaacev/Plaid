package lang

import (
	"fmt"
	"testing"
)

func TestScopeHasParent(t *testing.T) {
	root := MakeGlobalScope()
	child := MakeLocalScope(root, TypeFunction{})
	expectBool(t, root.HasParent(), false)
	expectBool(t, child.HasParent(), true)
}

func TestScopeErrors(t *testing.T) {
	root := MakeGlobalScope()
	child := MakeLocalScope(root, TypeFunction{})
	child.NewError(fmt.Errorf("foo bar baz"))
	expectNthError(t, root, 0, "foo bar baz")
	expectNthError(t, child, 0, "foo bar baz")
}

func TestScopeAddError(t *testing.T) {
	scope := MakeGlobalScope()
	expectNoScopeErrors(t, scope)
	scope.NewError(fmt.Errorf("a semantic analysis error"))
	expectNthError(t, scope, 0, "a semantic analysis error")
}

func TestScopeHasLocalVariable(t *testing.T) {
	scope := MakeGlobalScope()
	scope.NewVariable("foo", TypeIdent{Name: "Bar"})
	expectBool(t, scope.HasLocalVariable("foo"), true)
	expectBool(t, scope.HasLocalVariable("baz"), false)
}

func TestScopeNewVariable(t *testing.T) {
	scope := MakeGlobalScope()
	scope.NewVariable("foo", TypeIdent{Name: "Bar"})
	if scope.HasVariable("foo") {
		expectEquivalentType(t, scope.GetVariableType("foo"), TypeIdent{Name: "Bar"})
	} else {
		t.Errorf("Expected key '%s' in Scope#variables, none found", "foo")
	}
}

func TestScopeGetVariable(t *testing.T) {
	parent := MakeGlobalScope()
	child := MakeLocalScope(parent, TypeFunction{})
	parent.NewVariable("foo", TypeIdent{Name: "Bar"})
	expectEquivalentType(t, child.GetVariableType("foo"), TypeIdent{Name: "Bar"})
}

func TestScopeString(t *testing.T) {
	scope := MakeGlobalScope()
	scope.NewVariable("num", TypeIdent{Name: "Int"})
	scope.NewVariable("test", TypeIdent{Name: "Bool"})
	scope.NewVariable("coord", TypeTuple{Children: []Type{TypeIdent{Name: "Int"}, TypeIdent{Name: "Int"}}})

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
