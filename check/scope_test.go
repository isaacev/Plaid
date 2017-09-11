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
	scope.registerLocalVariable("foo", TypeIdent{"Bar"})
	expectBool(t, scope.hasLocalVariable("foo"), true)
	expectBool(t, scope.hasLocalVariable("baz"), false)
}

func TestScopeRegisterLocalVariable(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("foo", TypeIdent{"Bar"})
	typ, exists := scope.values["foo"]
	if exists {
		expectEquivalentType(t, typ, TypeIdent{"Bar"})
	} else {
		t.Errorf("Expected key '%s' in Scope#variables, none found", "foo")
	}
}

func TestScopeGetVariable(t *testing.T) {
	parent := makeScope(nil, nil)
	child := makeScope(parent, nil)
	parent.registerLocalVariable("foo", TypeIdent{"Bar"})
	expectEquivalentType(t, child.getVariable("foo"), TypeIdent{"Bar"})
	expectNil(t, child.getVariable("baz"))
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
	scope.registerLocalVariable("num", TypeIdent{"Int"})
	scope.registerLocalVariable("test", TypeIdent{"Bool"})
	scope.registerLocalVariable("coord", TypeTuple{[]Type{TypeIdent{"Int"}, TypeIdent{"Int"}}})

	expectString(t, scope.String(), "num : Int\ntest : Bool\ncoord : (Int Int)\n")
}
