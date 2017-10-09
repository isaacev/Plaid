package check

import (
	"plaid/lexer"
	"plaid/parser"
	"plaid/types"
	"plaid/vm"
	"testing"
)

var nop = lexer.Token{}

func TestCheckMain(t *testing.T) {
	scope := Check(parser.Program{})
	expectNoErrors(t, scope.Errors())

	var lib1 vm.Library = map[string]*vm.Builtin{
		"foo": &vm.Builtin{
			Type: types.Bool,
			Func: func(args []vm.Object) (vm.Object, error) { return nil, nil },
		},
	}
	scope = Check(parser.Program{}, lib1)
	expectBool(t, scope.getVariable("foo").Equals(types.Bool), true)
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

	prog, _ = parser.Parse("let a := fn ():Void { return 123; };")
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
	expectEquivalentType(t, scope.values["a"], types.Int)

	prog, _ = parser.Parse("let a := 1;")
	scope = Check(prog)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, scope.values["a"], types.Int)

	prog, _ = parser.Parse("let a := \"abc\";")
	scope = Check(prog)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, scope.values["a"], types.Str)

	prog, _ = parser.Parse("let a := fn () {};")
	scope = Check(prog)
	expectNoErrors(t, scope.Errors())

	prog, _ = parser.Parse("let a := true;")
	scope = Check(prog)
	expectNoErrors(t, scope.Errors())

	prog, _ = parser.Parse("let a := false;")
	scope = Check(prog)
	expectNoErrors(t, scope.Errors())

	prog, _ = parser.Parse("let a := add(2, 2);")
	scope = Check(prog)
	expectAnError(t, scope.errs[0], "variable 'add' was used before it was declared")
	expectBool(t, scope.values["a"].IsError(), true)

	prog, _ = parser.Parse("let f := fn():Void{}; let a := f();")
	scope = Check(prog)
	expectAnError(t, scope.errs[0], "cannot use void types in an expression")
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
	expectEquivalentType(t, scope.values["f"], types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{types.TypeIdent{Name: "Int"}}},
		Ret:    types.TypeIdent{Name: "Int"},
	})
}

func TestCheckDispatchExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("add", types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{
			types.TypeIdent{Name: "Int"},
			types.TypeIdent{Name: "Int"},
		}},
		Ret: types.TypeIdent{Name: "Int"},
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
	expectEquivalentType(t, typ, types.Int)

	scope = makeScope(nil, nil)
	scope.registerLocalVariable("add", types.Int)
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
	scope.registerLocalVariable("add", types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{
			types.TypeIdent{Name: "Int"},
			types.TypeIdent{Name: "Int"},
		}},
		Ret: types.TypeIdent{Name: "Int"},
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
	scope.registerLocalVariable("add", types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{
			types.TypeIdent{Name: "Int"},
			types.TypeIdent{Name: "Int"},
		}},
		Ret: types.TypeIdent{Name: "Int"},
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

func TestCheckAssignExpr(t *testing.T) {
	prog, _ := parser.Parse("a := 456;")
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("a", types.Int)
	checkProgram(scope, prog)
	expectNoErrors(t, scope.Errors())

	prog, _ = parser.Parse("a := 456;")
	scope = makeScope(nil, nil)
	scope.registerLocalVariable("a", types.Str)
	checkProgram(scope, prog)
	expectAnError(t, scope.Errors()[0], "'Str' cannot be assigned type 'Int'")

	prog, _ = parser.Parse("a := \"a\" + 45;")
	scope = makeScope(nil, nil)
	scope.registerLocalVariable("a", types.Str)
	checkProgram(scope, prog)
	expectAnError(t, scope.Errors()[0], "operator '+' does not support Str and Int")

	prog, _ = parser.Parse("a := 123;")
	scope = makeScope(nil, nil)
	checkProgram(scope, prog)
	expectAnError(t, scope.Errors()[0], "'a' cannot be assigned before it is declared")
}

func TestCheckBinaryExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("a", types.Int)
	scope.registerLocalVariable("b", types.Int)
	leftExpr := parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr := parser.IdentExpr{Tok: nop, Name: "b"}
	expr := parser.BinaryExpr{Tok: nop, Oper: "+", Left: leftExpr, Right: rightExpr}
	typ := checkBinaryExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, types.Int)

	scope = makeScope(nil, nil)
	scope.registerLocalVariable("a", types.Int)
	scope.registerLocalVariable("b", types.Int)
	leftExpr = parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr = parser.IdentExpr{Tok: nop, Name: "b"}
	expr = parser.BinaryExpr{Tok: nop, Oper: "-", Left: leftExpr, Right: rightExpr}
	typ = checkBinaryExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, types.Int)

	expr = parser.BinaryExpr{Tok: nop, Oper: "@", Left: leftExpr, Right: rightExpr}
	typ = checkBinaryExpr(scope, expr)
	expectAnError(t, scope.errs[0], "unknown infix operator '@'")
	expectBool(t, typ.IsError(), true)
}

func TestCheckIdentExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	scope.registerLocalVariable("x", types.Int)
	expr := parser.IdentExpr{Tok: nop, Name: "x"}
	typ := checkIdentExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, types.Int)

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
	expectEquivalentType(t, typ, types.Int)
}

func TestCheckStringExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	expr := parser.StringExpr{Tok: nop, Val: "abc"}
	typ := checkStringExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, types.Str)
}

func TestCheckBooleanExpr(t *testing.T) {
	scope := makeScope(nil, nil)
	expr := parser.BooleanExpr{Tok: nop, Val: true}
	typ := checkBooleanExpr(scope, expr)
	expectNoErrors(t, scope.Errors())
	expectEquivalentType(t, typ, types.Bool)
}

func TestConvertTypeSig(t *testing.T) {
	var note parser.TypeNote

	note = parser.TypeNoteVoid{Tok: nop}
	expectEquivalentType(t, types.ConvertTypeNote(note), types.TypeVoid{})

	note = parser.TypeNoteFunction{
		Params: parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
			parser.TypeNoteIdent{Tok: nop, Name: "Int"},
			parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
		}},
		Ret: parser.TypeNoteIdent{Tok: nop, Name: "Str"},
	}
	expectEquivalentType(t, types.ConvertTypeNote(note), types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{
			types.TypeIdent{Name: "Int"},
			types.TypeIdent{Name: "Bool"},
		}},
		Ret: types.TypeIdent{Name: "Str"},
	})

	note = parser.TypeNoteFunction{
		Params: parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
			parser.TypeNoteIdent{Tok: nop, Name: "Int"},
			parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
		}},
		Ret: parser.TypeNoteVoid{},
	}
	expectEquivalentType(t, types.ConvertTypeNote(note), types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{
			types.TypeIdent{Name: "Int"},
			types.TypeIdent{Name: "Bool"},
		}},
		Ret: types.TypeVoid{},
	})

	note = parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
		parser.TypeNoteIdent{Tok: nop, Name: "Int"},
		parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
	}}
	expectEquivalentType(t, types.ConvertTypeNote(note), types.TypeTuple{Children: []types.Type{
		types.TypeIdent{Name: "Int"},
		types.TypeIdent{Name: "Bool"},
	}})

	note = parser.TypeNoteList{Tok: nop, Child: parser.TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, types.ConvertTypeNote(note), types.TypeList{Child: types.TypeIdent{Name: "Int"}})

	note = parser.TypeNoteOptional{Tok: nop, Child: parser.TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, types.ConvertTypeNote(note), types.TypeOptional{Child: types.TypeIdent{Name: "Int"}})

	note = parser.TypeNoteIdent{Tok: nop, Name: "Int"}
	expectEquivalentType(t, types.ConvertTypeNote(note), types.TypeIdent{Name: "Int"})

	note = nil
	expectBool(t, types.ConvertTypeNote(note) == nil, true)
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

func expectEquivalentType(t *testing.T, t1 types.Type, t2 types.Type) {
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
