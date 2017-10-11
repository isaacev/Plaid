package check

import (
	"fmt"
	"plaid/types"
	"testing"
)

func TestScopeHasParent(t *testing.T) {
	root := makeScope(nil, nil)
	child := makeScope(root, nil)
	expectBool(t, root.hasParent(), false)
	expectBool(t, child.hasParent(), true)
}

func TestScopeErrors(t *testing.T) {
	root := makeScope(nil, nil)
	child := makeScope(root, nil)
	child.addError(fmt.Errorf("foo bar baz"))
	expectNoErrors(t, child.errs)
	expectAnError(t, root.errs[0], "foo bar baz")
	expectAnError(t, child.Errors()[0], "foo bar baz")
}

func TestScopeAddError(t *testing.T) {
	scope := makeScope(nil, nil)
	expectNoErrors(t, scope.Errors())
	scope.addError(fmt.Errorf("a semantic analysis error"))
	expectAnError(t, scope.Errors()[0], "a semantic analysis error")
}

func TestScopeHasLocalVariable(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("foo", types.TypeIdent{Name: "Bar"})
	expectBool(t, scope.hasLocalVariable("foo"), true)
	expectBool(t, scope.hasLocalVariable("baz"), false)
}

func TestScoperegisterLocalVariable(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("foo", types.TypeIdent{Name: "Bar"})
	typ, exists := scope.Values["foo"]
	if exists {
		expectEquivalentType(t, typ, types.TypeIdent{Name: "Bar"})
	} else {
		t.Errorf("Expected key '%s' in Scope#variables, none found", "foo")
	}
}

func TestScopeGetVariable(t *testing.T) {
	parent := makeScope(nil, nil)
	child := makeScope(parent, nil)
	parent.registerLocalVariable("foo", types.TypeIdent{Name: "Bar"})
	expectEquivalentType(t, child.getVariable("foo"), types.TypeIdent{Name: "Bar"})
	expectNil(t, child.getVariable("baz"))
}

func TestScopeString(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("num", types.TypeIdent{Name: "Int"})
	scope.registerLocalVariable("test", types.TypeIdent{Name: "Bool"})
	scope.registerLocalVariable("coord", types.TypeTuple{Children: []types.Type{types.TypeIdent{Name: "Int"}, types.TypeIdent{Name: "Int"}}})

	expectString(t, scope.String(), "coord : (Int Int)\nnum   : Int\ntest  : Bool")
}
