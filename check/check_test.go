package check

import (
	"plaid/lexer"
	"plaid/parser"
	"testing"
)

var nop = lexer.Token{}

func TestCheckMain(t *testing.T) {
	scope := Check(parser.Program{})
	expectNoErrors(t, scope.Errors())
}

func TestCheckProgram(t *testing.T) {
	prog, _ := parser.Parse("let a := 123;")
	scope := makeScope(nil, nil)
	checkProgram(scope, prog)
	expectNoErrors(t, scope.Errors())
}

func TestCheckStmt(t *testing.T) {
	prog, _ := parser.Parse("let a := 123;")
	scope := makeScope(nil, nil)
	checkStmt(scope, prog.Stmts[0])
	expectNoErrors(t, scope.Errors())
}

func TestCheckReturnStmt(t *testing.T) {
	prog, _ := parser.Parse("let a := fn (): Int { return \"abc\"; };")
	scope := Check(prog)
	expectAnError(t, scope.errs[0], "expected to return 'Int', got 'Str'")

	prog, _ = parser.Parse("let a := fn (): Int { return x; };")
	scope = Check(prog)
	expectAnError(t, scope.errs[0], "variable 'x' was used before it was declared")

	prog, _ = parser.Parse("let a := fn (): Int { return; };")
	scope = Check(prog)
	expectAnError(t, scope.errs[0], "expected a return type of 'Int', got nothing")

	prog, _ = parser.Parse("let a := fn () { return 123; };")
	scope = Check(prog)
	expectAnError(t, scope.errs[0], "expected to return nothing, got 'Int'")

	prog, _ = parser.Parse("return;")
	scope = Check(prog)
	expectAnError(t, scope.errs[0], "return statements must be inside a function")
}

func TestCheckExpr(t *testing.T) {
	prog, _ := parser.Parse("let a := 2 + 1;")
	scope := Check(prog)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, scope.values["a"], BuiltinInt)

	prog, _ = parser.Parse("let a := 1;")
	scope = Check(prog)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, scope.values["a"], BuiltinInt)

	prog, _ = parser.Parse("let a := \"abc\";")
	scope = Check(prog)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, scope.values["a"], BuiltinStr)

	prog, _ = parser.Parse("let a := fn () {};")
	scope = Check(prog)
	expectNoErrors(t, scope.Errors())

	prog, _ = parser.Parse("let a := add(2, 2);")
	scope = Check(prog)
	expectAnError(t, scope.errs[0], "variable 'add' was used before it was declared")
	expectBool(t, scope.values["a"].IsError(), true)

	prog, _ = parser.Parse("let a := -5;")
	scope = Check(prog)
	expectAnError(t, scope.errs[0], "unknown expression type")
	expectBool(t, scope.values["a"].IsError(), true)
}

func TestCheckFunctionExpr(t *testing.T) {
	prog, _ := parser.Parse("let f := fn (a: Int): Int { };")
	scope := Check(prog)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, scope.values["f"], TypeFunction{
		TypeTuple{[]Type{TypeIdent{"Int"}}},
		TypeIdent{"Int"},
	})
}

func TestCheckDispatchExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("add", TypeFunction{
		TypeTuple{[]Type{
			TypeIdent{"Int"},
			TypeIdent{"Int"},
		}},
		TypeIdent{"Int"},
	})
	expr := parser.DispatchExpr{
		Callee: parser.IdentExpr{Tok: nop, Name: "add"},
		Args: []parser.Expr{
			parser.NumberExpr{Tok: nop, Val: 2},
			parser.NumberExpr{Tok: nop, Val: 5},
		},
	}
	typ := checkDispatchExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, BuiltinInt)

	scope = makeScope(nil, nil)
	scope.registerLocalVariable("add", BuiltinInt)
	expr = parser.DispatchExpr{
		Callee: parser.IdentExpr{Tok: nop, Name: "add"},
		Args: []parser.Expr{
			parser.NumberExpr{Tok: nop, Val: 2},
			parser.NumberExpr{Tok: nop, Val: 5},
		},
	}
	typ = checkDispatchExpr(scope, expr)
	expectAnError(t, scope.errs[0], "cannot call function on type 'Int'")
	expectBool(t, typ.IsError(), true)

	scope = makeScope(nil, nil)
	scope.registerLocalVariable("add", TypeFunction{
		TypeTuple{[]Type{
			TypeIdent{"Int"},
			TypeIdent{"Int"},
		}},
		TypeIdent{"Int"},
	})
	expr = parser.DispatchExpr{
		Callee: parser.IdentExpr{Tok: nop, Name: "add"},
		Args: []parser.Expr{
			parser.NumberExpr{Tok: nop, Val: 2},
		},
	}
	typ = checkDispatchExpr(scope, expr)
	expectAnError(t, scope.errs[0], "expected 2 arguments, got 1")
	expectBool(t, typ.IsError(), true)

	scope = makeScope(nil, nil)
	scope.registerLocalVariable("add", TypeFunction{
		TypeTuple{[]Type{
			TypeIdent{"Int"},
			TypeIdent{"Int"},
		}},
		TypeIdent{"Int"},
	})
	expr = parser.DispatchExpr{
		Callee: parser.IdentExpr{Tok: nop, Name: "add"},
		Args: []parser.Expr{
			parser.StringExpr{Tok: nop, Val: "2"},
			parser.StringExpr{Tok: nop, Val: "4"},
		},
	}
	typ = checkDispatchExpr(scope, expr)
	expectAnError(t, scope.errs[0], "expected 'Int', got 'Str'")
	expectAnError(t, scope.errs[1], "expected 'Int', got 'Str'")
	expectBool(t, typ.IsError(), true)
}

func TestCheckBinaryExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("a", BuiltinInt)
	scope.registerLocalVariable("b", BuiltinInt)
	leftExpr := parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr := parser.IdentExpr{Tok: nop, Name: "b"}
	expr := parser.BinaryExpr{Tok: nop, Oper: "+", Left: leftExpr, Right: rightExpr}
	typ := checkBinaryExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, BuiltinInt)

	expr = parser.BinaryExpr{Tok: nop, Oper: "@", Left: leftExpr, Right: rightExpr}
	typ = checkBinaryExpr(scope, expr)
	expectAnError(t, scope.errs[0], "unknown infix operator '@'")
	expectBool(t, typ.IsError(), true)
}

func TestCheckAddition(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("a", BuiltinInt)
	scope.registerLocalVariable("b", BuiltinInt)
	leftExpr := parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr := parser.IdentExpr{Tok: nop, Name: "b"}
	typ := checkAddition(scope, leftExpr, rightExpr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, BuiltinInt)

	scope = makeScope(nil, nil)
	scope.registerLocalVariable("a", BuiltinStr)
	scope.registerLocalVariable("b", BuiltinInt)
	leftExpr = parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr = parser.IdentExpr{Tok: nop, Name: "b"}
	typ = checkAddition(scope, leftExpr, rightExpr)
	expectAnError(t, scope.errs[0], "left side must have type Int, got Str")
	expectBool(t, typ.IsError(), true)

	scope = makeScope(nil, nil)
	scope.registerLocalVariable("a", BuiltinInt)
	scope.registerLocalVariable("b", BuiltinStr)
	leftExpr = parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr = parser.IdentExpr{Tok: nop, Name: "b"}
	typ = checkAddition(scope, leftExpr, rightExpr)
	expectAnError(t, scope.errs[0], "right side must have type Int, got Str")
	expectBool(t, typ.IsError(), true)

	scope = makeScope(nil, nil)
	scope.registerLocalVariable("a", TypeError{})
	scope.registerLocalVariable("b", BuiltinStr)
	leftExpr = parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr = parser.IdentExpr{Tok: nop, Name: "b"}
	typ = checkAddition(scope, leftExpr, rightExpr)
	expectNoErrors(t, scope.Errors())
	expectBool(t, typ.IsError(), true)
}

func TestCheckIdentExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("x", BuiltinInt)
	expr := parser.IdentExpr{Tok: nop, Name: "x"}
	typ := checkIdentExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, BuiltinInt)

	scope = makeScope(nil, nil)
	expr = parser.IdentExpr{Tok: nop, Name: "x"}
	typ = checkIdentExpr(scope, expr)
	expectAnError(t, scope.errs[0], "variable 'x' was used before it was declared")
	expectBool(t, typ.IsError(), true)
}

func TestCheckNumberExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	expr := parser.NumberExpr{Tok: nop, Val: 123}
	typ := checkNumberExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, BuiltinInt)
}

func TestCheckStringExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	expr := parser.StringExpr{Tok: nop, Val: "abc"}
	typ := checkStringExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, BuiltinStr)
}

func TestConvertTypeSig(t *testing.T) {
	var note parser.TypeNote

	note = parser.TypeNoteFunction{
		Params: parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
			parser.TypeNoteIdent{Tok: nop, Name: "Int"},
			parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
		}},
		Ret: parser.TypeNoteIdent{Tok: nop, Name: "Str"},
	}
	expectEquivalentType(t, convertTypeNote(note), TypeFunction{
		TypeTuple{[]Type{
			TypeIdent{"Int"},
			TypeIdent{"Bool"},
		}},
		TypeIdent{"Str"},
	})

	note = parser.TypeNoteFunction{
		Params: parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
			parser.TypeNoteIdent{Tok: nop, Name: "Int"},
			parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
		}},
		Ret: nil,
	}
	expectEquivalentType(t, convertTypeNote(note), TypeFunction{
		TypeTuple{[]Type{
			TypeIdent{"Int"},
			TypeIdent{"Bool"},
		}},
		nil,
	})

	note = parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
		parser.TypeNoteIdent{Tok: nop, Name: "Int"},
		parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
	}}
	expectEquivalentType(t, convertTypeNote(note), TypeTuple{[]Type{
		TypeIdent{"Int"},
		TypeIdent{"Bool"},
	}})

	note = parser.TypeNoteList{Tok: nop, Child: parser.TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, convertTypeNote(note), TypeList{TypeIdent{"Int"}})

	note = parser.TypeNoteOptional{Tok: nop, Child: parser.TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, convertTypeNote(note), TypeOptional{TypeIdent{"Int"}})

	note = parser.TypeNoteIdent{Tok: nop, Name: "Int"}
	expectEquivalentType(t, convertTypeNote(note), TypeIdent{"Int"})

	note = nil
	expectBool(t, convertTypeNote(note) == nil, true)
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
