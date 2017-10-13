package scope

import (
	"fmt"
	"plaid/types"
	"testing"
)

func TestScopeHasParent(t *testing.T) {
	root := MakeGlobalScope()
	child := MakeLocalScope(root, types.Function{})
	expectBool(t, root.HasParent(), false)
	expectBool(t, child.HasParent(), true)
}

func TestScopeErrors(t *testing.T) {
	root := MakeGlobalScope()
	child := MakeLocalScope(root, types.Function{})
	child.NewError(fmt.Errorf("foo bar baz"))
	expectNthError(t, root, 0, "foo bar baz")
	expectNthError(t, child, 0, "foo bar baz")
}

func TestScopeAddError(t *testing.T) {
	scope := MakeGlobalScope()
	expectNoErrors(t, scope)
	scope.NewError(fmt.Errorf("a semantic analysis error"))
	expectNthError(t, scope, 0, "a semantic analysis error")
}

func TestScopeHasLocalVariable(t *testing.T) {
	scope := MakeGlobalScope()
	scope.NewVariable("foo", types.Ident{Name: "Bar"})
	expectBool(t, scope.HasLocalVariable("foo"), true)
	expectBool(t, scope.HasLocalVariable("baz"), false)
}

func TestScopeNewVariable(t *testing.T) {
	scope := MakeGlobalScope()
	scope.NewVariable("foo", types.Ident{Name: "Bar"})
	if scope.HasVariable("foo") {
		expectEquivalentType(t, scope.GetVariableType("foo"), types.Ident{Name: "Bar"})
	} else {
		t.Errorf("Expected key '%s' in Scope#variables, none found", "foo")
	}
}

func TestScopeGetVariable(t *testing.T) {
	parent := MakeGlobalScope()
	child := MakeLocalScope(parent, types.Function{})
	parent.NewVariable("foo", types.Ident{Name: "Bar"})
	expectEquivalentType(t, child.GetVariableType("foo"), types.Ident{Name: "Bar"})
}

func TestScopeString(t *testing.T) {
	scope := MakeGlobalScope()
	scope.NewVariable("num", types.Ident{Name: "Int"})
	scope.NewVariable("test", types.Ident{Name: "Bool"})
	scope.NewVariable("coord", types.Tuple{Children: []types.Type{types.Ident{Name: "Int"}, types.Ident{Name: "Int"}}})

	expectString(t, scope.String(), "coord : (Int Int)\nnum   : Int\ntest  : Bool")
}

func expectNthError(t *testing.T, scope Scope, n int, msg string) {
	if len(scope.GetErrors()) <= n {
		t.Fatalf("Expected at least %d errors", n+1)
	}

	expectAnError(t, scope.GetErrors()[n], msg)
}

func expectAnError(t *testing.T, err error, msg string) {
	if err == nil {
		t.Errorf("Expected an error '%s', got no errors", err)
	} else if msg != err.Error() {
		t.Errorf("Expected '%s', got '%s'", msg, err)
	}
}

func expectNoErrors(t *testing.T, scope Scope) {
	if scope.HasErrors() {
		for i, err := range scope.GetErrors() {
			t.Errorf("%d '%s'", i, err)
		}

		t.Fatalf("Expected no errors, found %d", len(scope.GetErrors()))
	}
}

func expectEquivalentType(t *testing.T, t1 types.Type, t2 types.Type) {
	if t1 == nil {
		t.Fatalf("Expected type, got <nil>")
	}

	if t2 == nil {
		t.Fatalf("Expected type, got <nil>")
	}

	same := t1.Equals(t2)
	commutative := t1.Equals(t2) == t2.Equals(t1)

	if commutative == false {
		if same {
			t.Errorf("%s == %s, but %s != %s", t1, t2, t2, t1)
		} else {
			t.Errorf("%s == %s, but %s != %s", t2, t1, t1, t2)
		}
	}

	if same == false {
		t.Errorf("Expected %s == %s, got %t", t1, t2, same)
	}
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
