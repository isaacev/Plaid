package check

import (
	"fmt"
	"testing"
)

func TestScopeHasParent(t *testing.T) {
	root := makeScope(nil, nil)
	child := makeScope(root, nil)
	expectBool(t, root.hasParent(), false)
	expectBool(t, child.hasParent(), true)
}

func TestScopeAddError(t *testing.T) {
	scope := makeScope(nil, nil)
	expectNoErrors(t, scope.Errors())
	scope.addError(fmt.Errorf("a semantic analysis error"))
	expectAnError(t, scope.errs[0], "a semantic analysis error")

	root := makeScope(nil, nil)
	child := makeScope(root, nil)
	expectNoErrors(t, root.Errors())
	expectNoErrors(t, child.Errors())
	child.addError(fmt.Errorf("a semantic analysis error"))
	expectNoErrors(t, child.Errors())
	expectAnError(t, root.errs[0], "a semantic analysis error")
}

func TestScopeHasVariable(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerVariable("foo", TypeIdent{"Bar"})
	expectBool(t, scope.hasVariable("foo"), true)
	expectBool(t, scope.hasVariable("baz"), false)
}

func TestScopeRegisterVariable(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerVariable("foo", TypeIdent{"Bar"})
	typ, exists := scope.values["foo"]
	if exists {
		expectEquivalentType(t, typ, TypeIdent{"Bar"})
	} else {
		t.Errorf("Expected key '%s' in Scope#variables, none found", "foo")
	}
}

func TestScopeGetVariable(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerVariable("foo", TypeIdent{"Bar"})
	expectEquivalentType(t, scope.getVariable("foo"), TypeIdent{"Bar"})
	expectNil(t, scope.getVariable("baz"))
}

func TestScopeHasPendingReturnType(t *testing.T) {
	scope := makeScope(nil, nil)
	expectBool(t, scope.hasPendingReturnType(), false)

	scope = makeScope(nil, TypeIdent{"Int"})
	expectBool(t, scope.hasPendingReturnType(), true)
}

func TestScopeGetPendingReturnType(t *testing.T) {
	scope := makeScope(nil, nil)
	expectBool(t, scope.getPendingReturnType() == nil, true)

	scope = makeScope(nil, TypeIdent{"Int"})
	expectEquivalentType(t, scope.getPendingReturnType(), TypeIdent{"Int"})
}

func TestScopeSetPendingReturnType(t *testing.T) {
	scope := makeScope(nil, nil)
	expectBool(t, scope.hasPendingReturnType(), false)

	scope.setPendingReturnType(TypeIdent{"Int"})
	expectBool(t, scope.pendingReturn.Equals(TypeIdent{"Int"}), true)
	expectBool(t, scope.hasPendingReturnType(), true)
}

func TestScopeString(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerVariable("num", TypeIdent{"Int"})
	scope.registerVariable("test", TypeIdent{"Bool"})
	scope.registerVariable("coord", TypeTuple{[]Type{TypeIdent{"Int"}, TypeIdent{"Int"}}})

	expectString(t, scope.String(), "num : Int\ntest : Bool\ncoord : (Int Int)\n")
}
