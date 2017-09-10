package check

import (
	"fmt"
	"plaid/lexer"
	"plaid/parser"
	"testing"
)

var nop = lexer.Token{}

func TestCheckMain(t *testing.T) {
	scope := Check(parser.Program{})
	expectNoErrors(t, scope.Errs)
}

func TestScopeHasParent(t *testing.T) {
	root := makeScope(nil)
	child := makeScope(root)
	expectBool(t, root.hasParent(), false)
	expectBool(t, child.hasParent(), true)
}

func TestScopeRegisterVariable(t *testing.T) {
	scope := makeScope(nil)
	scope.registerVariable("foo", TypeIdent{"Bar"})
	typ, exists := scope.variables["foo"]
	if exists {
		expectEquivalentType(t, typ, TypeIdent{"Bar"})
	} else {
		t.Errorf("Expected key '%s' in Scope#variables, none found", "foo")
	}
}

func TestScopeHasVariable(t *testing.T) {
	scope := makeScope(nil)
	scope.registerVariable("foo", TypeIdent{"Bar"})
	expectBool(t, scope.hasVariable("foo"), true)
	expectBool(t, scope.hasVariable("baz"), false)
}

func TestScopeGetVariable(t *testing.T) {
	scope := makeScope(nil)
	scope.registerVariable("foo", TypeIdent{"Bar"})
	expectEquivalentType(t, scope.getVariable("foo"), TypeIdent{"Bar"})
	expectNil(t, scope.getVariable("baz"))
}

func TestScopeAddError(t *testing.T) {
	scope := makeScope(nil)
	expectNoErrors(t, scope.Errs)
	scope.addError(fmt.Errorf("a semantic analysis error"))
	expectAnError(t, scope.Errs[0], "a semantic analysis error")

	root := makeScope(nil)
	child := makeScope(root)
	expectNoErrors(t, root.Errs)
	expectNoErrors(t, child.Errs)
	child.addError(fmt.Errorf("a semantic analysis error"))
	expectNoErrors(t, child.Errs)
	expectAnError(t, root.Errs[0], "a semantic analysis error")
}

func TestScopeString(t *testing.T) {
	scope := makeScope(nil)
	scope.registerVariable("num", TypeIdent{"Int"})
	scope.registerVariable("test", TypeIdent{"Bool"})
	scope.registerVariable("coord", TypeTuple{[]Type{TypeIdent{"Int"}, TypeIdent{"Int"}}})

	expectString(t, scope.String(), `+----------+--------------+
| Var      | Type         |
| -------- | ------------ |
| coord    | (Int Int)    |
| num      | Int          |
| test     | Bool         |
+----------+--------------+
`)
}

func TestCheckProgram(t *testing.T) {
	prog, _ := parser.Parse("let a := 123;")
	scope := makeScope(nil)
	checkProgram(scope, prog)
	expectNoErrors(t, scope.Errs)
}

func TestCheckAddition(t *testing.T) {
	scope := makeScope(nil)
	scope.registerVariable("a", BuiltinInt)
	scope.registerVariable("b", BuiltinInt)
	leftExpr := parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr := parser.IdentExpr{Tok: nop, Name: "b"}
	typ := checkAddition(scope, leftExpr, rightExpr)
	expectNoErrors(t, scope.Errs)
	expectEquivalentType(t, typ, BuiltinInt)

	scope = makeScope(nil)
	scope.registerVariable("a", BuiltinStr)
	scope.registerVariable("b", BuiltinInt)
	leftExpr = parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr = parser.IdentExpr{Tok: nop, Name: "b"}
	typ = checkAddition(scope, leftExpr, rightExpr)
	expectAnError(t, scope.Errs[0], "left side must have type Int, got Str")
	expectBool(t, typ.IsError(), true)

	scope = makeScope(nil)
	scope.registerVariable("a", BuiltinInt)
	scope.registerVariable("b", BuiltinStr)
	leftExpr = parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr = parser.IdentExpr{Tok: nop, Name: "b"}
	typ = checkAddition(scope, leftExpr, rightExpr)
	expectAnError(t, scope.Errs[0], "right side must have type Int, got Str")
	expectBool(t, typ.IsError(), true)

	scope = makeScope(nil)
	scope.registerVariable("a", TypeError{})
	scope.registerVariable("b", BuiltinStr)
	leftExpr = parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr = parser.IdentExpr{Tok: nop, Name: "b"}
	typ = checkAddition(scope, leftExpr, rightExpr)
	expectNoErrors(t, scope.Errs)
	expectBool(t, typ.IsError(), true)
}

func TestCheckIdentExpr(t *testing.T) {
	scope := makeScope(nil)
	scope.registerVariable("x", BuiltinInt)
	expr := parser.IdentExpr{Tok: nop, Name: "x"}
	typ := checkIdentExpr(scope, expr)
	expectNoErrors(t, scope.Errs)
	expectEquivalentType(t, typ, BuiltinInt)

	scope = makeScope(nil)
	expr = parser.IdentExpr{Tok: nop, Name: "x"}
	typ = checkIdentExpr(scope, expr)
	expectAnError(t, scope.Errs[0], "variable 'x' was used before it was declared")
	expectBool(t, typ.IsError(), true)
}

func TestCheckNumberExpr(t *testing.T) {
	scope := makeScope(nil)
	expr := parser.NumberExpr{Tok: nop, Val: 123}
	typ := checkNumberExpr(scope, expr)
	expectNoErrors(t, scope.Errs)
	expectEquivalentType(t, typ, BuiltinInt)
}

func TestCheckStringExpr(t *testing.T) {
	scope := makeScope(nil)
	expr := parser.StringExpr{Tok: nop, Val: "abc"}
	typ := checkStringExpr(scope, expr)
	expectNoErrors(t, scope.Errs)
	expectEquivalentType(t, typ, BuiltinStr)
}

func expectNoErrors(t *testing.T, errs []error) {
	if len(errs) > 0 {
		for i, err := range errs {
			t.Errorf("%d '%s'", i, err)
		}

		t.Fatalf("Expected no errors, found %d", len(errs))
	}
}

func expectAnError(t *testing.T, err error, msg string) {
	if err == nil {
		t.Errorf("Expected an error '%s', got no errors", err)
	} else if msg != err.Error() {
		t.Errorf("Expected '%s', got '%s'", msg, err)
	}
}

func expectNil(t *testing.T, got interface{}) {
	if got != nil {
		t.Errorf("Expected nil, got '%v'", got)
	}
}
