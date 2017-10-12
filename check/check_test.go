package check

import (
	"plaid/lexer"
	"plaid/parser"
	"plaid/scope"
	"plaid/types"
	"plaid/vm"
	"testing"
)

var nop = lexer.Token{}

func TestCheckMain(t *testing.T) {
	s := Check(parser.Program{})
	expectNoErrors(t, s)

	var lib1 vm.Library = map[string]*vm.Builtin{
		"foo": &vm.Builtin{
			Type: types.Bool,
			Func: func(args []vm.Object) (vm.Object, error) { return nil, nil },
		},
	}
	s = Check(parser.Program{}, lib1)
	expectLocalVariableType(t, s, "foo", types.Bool)
	expectNoErrors(t, s)
}

func TestCheckProgram(t *testing.T) {
	prog, _ := parser.Parse("let a := 123;")
	s := scope.MakeGlobalScope()
	checkProgram(s, prog)
	expectNoErrors(t, s)
}

func TestCheckStmt(t *testing.T) {
	prog, _ := parser.Parse("let a := 123;")
	s := scope.MakeGlobalScope()
	checkStmt(s, prog.Stmts[0])
	expectNoErrors(t, s)
}

func TestCheckIfStmt(t *testing.T) {
	prog, _ := parser.Parse("if true {};")
	s := Check(prog)
	expectNoErrors(t, s)

	prog, _ = parser.Parse("if 123 {};")
	s = Check(prog)
	expectNthError(t, s, 0, "condition must resolve to a boolean")
}

func TestCheckReturnStmt(t *testing.T) {
	prog, _ := parser.Parse("let a := fn (): Int { return \"abc\"; };")
	s := Check(prog)
	expectNthError(t, s, 0, "expected to return 'Int', got 'Str'")

	prog, _ = parser.Parse("let a := fn (): Int { return x; };")
	s = Check(prog)
	expectNthError(t, s, 0, "variable 'x' was used before it was declared")

	prog, _ = parser.Parse("let a := fn (): Int { return; };")
	s = Check(prog)
	expectNthError(t, s, 0, "expected a return type of 'Int', got nothing")

	prog, _ = parser.Parse("let a := fn ():Void { return 123; };")
	s = Check(prog)
	expectNthError(t, s, 0, "expected to return nothing, got 'Int'")

	prog, _ = parser.Parse("return;")
	s = Check(prog)
	expectNthError(t, s, 0, "return statements must be inside a function")
}

func TestCheckExpr(t *testing.T) {
	prog, _ := parser.Parse("let a := 2 + 1;")
	s := Check(prog)
	expectNoErrors(t, s)
	expectEquivalentType(t, s.GetVariableType("a"), types.Int)

	prog, _ = parser.Parse("let a := 1;")
	s = Check(prog)
	expectNoErrors(t, s)
	expectEquivalentType(t, s.GetVariableType("a"), types.Int)

	prog, _ = parser.Parse("let a := \"abc\";")
	s = Check(prog)
	expectNoErrors(t, s)
	expectEquivalentType(t, s.GetVariableType("a"), types.Str)

	prog, _ = parser.Parse("let a := fn () {};")
	s = Check(prog)
	expectNoErrors(t, s)

	prog, _ = parser.Parse("let a := true;")
	s = Check(prog)
	expectNoErrors(t, s)

	prog, _ = parser.Parse("let a := false;")
	s = Check(prog)
	expectNoErrors(t, s)

	prog, _ = parser.Parse("let a := [1, 2, 3];")
	s = Check(prog)
	expectNoErrors(t, s)

	prog, _ = parser.Parse("let a := \"abc\"[0];")
	s = Check(prog)
	expectNoErrors(t, s)

	prog, _ = parser.Parse("let a := add(2, 2);")
	s = Check(prog)
	expectNthError(t, s, 0, "variable 'add' was used before it was declared")
	expectBool(t, s.GetVariableType("a").IsError(), true)

	prog, _ = parser.Parse("let f := fn():Void{}; let a := f();")
	s = Check(prog)
	expectNthError(t, s, 0, "cannot use void types in an expression")
	expectBool(t, s.GetVariableType("a").IsError(), true)

	prog, _ = parser.Parse("let a := -5;")
	s = Check(prog)
	expectNthError(t, s, 0, "unknown expression type")
	expectBool(t, s.GetVariableType("a").IsError(), true)
}

func TestCheckFunctionExpr(t *testing.T) {
	prog, _ := parser.Parse("let f := fn (a: Int): Int { };")
	s := Check(prog)
	expectNoErrors(t, s)
	expectEquivalentType(t, s.GetVariableType("f"), types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{types.TypeIdent{Name: "Int"}}},
		Ret:    types.TypeIdent{Name: "Int"},
	})
}

func TestCheckDispatchExpr(t *testing.T) {
	s := scope.MakeGlobalScope()
	s.NewVariable("add", types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{
			types.TypeIdent{Name: "Int"},
			types.TypeIdent{Name: "Int"},
		}},
		Ret: types.TypeIdent{Name: "Int"},
	})
	expr := &parser.DispatchExpr{
		Callee: &parser.IdentExpr{Tok: nop, Name: "add"},
		Args: []parser.Expr{
			&parser.NumberExpr{Tok: nop, Val: 2},
			&parser.NumberExpr{Tok: nop, Val: 5},
		},
	}
	typ := checkDispatchExpr(s, expr)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Int)

	s = scope.MakeGlobalScope()
	s.NewVariable("add", types.Int)
	expr = &parser.DispatchExpr{
		Callee: &parser.IdentExpr{Tok: nop, Name: "add"},
		Args: []parser.Expr{
			&parser.NumberExpr{Tok: nop, Val: 2},
			&parser.NumberExpr{Tok: nop, Val: 5},
		},
	}
	typ = checkDispatchExpr(s, expr)
	expectNthError(t, s, 0, "cannot call function on type 'Int'")
	expectBool(t, typ.IsError(), true)

	s = scope.MakeGlobalScope()
	s.NewVariable("add", types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{
			types.TypeIdent{Name: "Int"},
			types.TypeIdent{Name: "Int"},
		}},
		Ret: types.TypeIdent{Name: "Int"},
	})
	expr = &parser.DispatchExpr{
		Callee: &parser.IdentExpr{Tok: nop, Name: "add"},
		Args: []parser.Expr{
			&parser.NumberExpr{Tok: nop, Val: 2},
		},
	}
	typ = checkDispatchExpr(s, expr)
	expectNthError(t, s, 0, "expected 2 arguments, got 1")
	expectBool(t, typ.IsError(), true)

	s = scope.MakeGlobalScope()
	s.NewVariable("foo", types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{
			types.TypeIdent{Name: "Int"},
		}},
		Ret: types.TypeIdent{Name: "Int"},
	})
	expr = &parser.DispatchExpr{
		Callee: &parser.IdentExpr{Tok: nop, Name: "foo"},
		Args: []parser.Expr{
			&parser.SelfExpr{Tok: nop},
		},
	}
	typ = checkDispatchExpr(s, expr)
	expectNthError(t, s, 0, "self references must be inside a function")
	expectBool(t, typ.IsError(), true)

	s = scope.MakeGlobalScope()
	s.NewVariable("add", types.TypeFunction{
		Params: types.TypeTuple{Children: []types.Type{
			types.TypeIdent{Name: "Int"},
			types.TypeIdent{Name: "Int"},
		}},
		Ret: types.TypeIdent{Name: "Int"},
	})
	expr = &parser.DispatchExpr{
		Callee: &parser.IdentExpr{Tok: nop, Name: "add"},
		Args: []parser.Expr{
			&parser.StringExpr{Tok: nop, Val: "2"},
			&parser.StringExpr{Tok: nop, Val: "4"},
		},
	}
	typ = checkDispatchExpr(s, expr)
	expectNthError(t, s, 0, "expected 'Int', got 'Str'")
	expectNthError(t, s, 1, "expected 'Int', got 'Str'")
	expectBool(t, typ.IsError(), true)
}

func TestCheckAssignExpr(t *testing.T) {
	prog, _ := parser.Parse("a := 456;")
	s := scope.MakeGlobalScope()
	s.NewVariable("a", types.Int)
	checkProgram(s, prog)
	expectNoErrors(t, s)

	prog, _ = parser.Parse("a := 456;")
	s = scope.MakeGlobalScope()
	s.NewVariable("a", types.Str)
	checkProgram(s, prog)
	expectNthError(t, s, 0, "'Str' cannot be assigned type 'Int'")

	prog, _ = parser.Parse("a := \"a\" + 45;")
	s = scope.MakeGlobalScope()
	s.NewVariable("a", types.Str)
	checkProgram(s, prog)
	expectNthError(t, s, 0, "operator '+' does not support Str and Int")

	prog, _ = parser.Parse("a := 123;")
	s = scope.MakeGlobalScope()
	checkProgram(s, prog)
	expectNthError(t, s, 0, "'a' cannot be assigned before it is declared")
}

func TestCheckBinaryExpr(t *testing.T) {
	s := scope.MakeGlobalScope()
	s.NewVariable("a", types.Int)
	s.NewVariable("b", types.Int)
	leftExpr := &parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr := &parser.IdentExpr{Tok: nop, Name: "b"}
	expr := &parser.BinaryExpr{Tok: nop, Oper: "+", Left: leftExpr, Right: rightExpr}
	typ := checkBinaryExpr(s, expr, defaultBinopsLUT)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Int)

	s = scope.MakeGlobalScope()
	s.NewVariable("a", types.Int)
	s.NewVariable("b", types.Int)
	leftExpr = &parser.IdentExpr{Tok: nop, Name: "a"}
	rightExpr = &parser.IdentExpr{Tok: nop, Name: "b"}
	expr = &parser.BinaryExpr{Tok: nop, Oper: "-", Left: leftExpr, Right: rightExpr}
	typ = checkBinaryExpr(s, expr, defaultBinopsLUT)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Int)

	s = scope.MakeGlobalScope()
	expr = &parser.BinaryExpr{Tok: nop, Oper: "+", Left: leftExpr, Right: rightExpr}
	typ = checkBinaryExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "variable 'a' was used before it was declared")
	expectNthError(t, s, 1, "variable 'b' was used before it was declared")
	expectBool(t, typ.IsError(), true)

	s = scope.MakeGlobalScope()
	s.NewVariable("a", types.Int)
	s.NewVariable("b", types.Int)
	expr = &parser.BinaryExpr{Tok: nop, Oper: "@", Left: leftExpr, Right: rightExpr}
	typ = checkBinaryExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "unknown infix operator '@'")
	expectBool(t, typ.IsError(), true)
}

func TestCheckListExpr(t *testing.T) {
	good := func(expr *parser.ListExpr, exp types.Type) {
		s := scope.MakeGlobalScope()
		got := checkListExpr(s, expr)
		expectNoErrors(t, s)
		expectEquivalentType(t, got, exp)
	}

	bad := func(expr *parser.ListExpr, exp string) {
		s := scope.MakeGlobalScope()
		got := checkListExpr(s, expr)
		expectNthError(t, s, 0, exp)
		expectEquivalentType(t, got, types.TypeError{})
	}

	good(&parser.ListExpr{Elements: []parser.Expr{
		&parser.StringExpr{Val: "foo"},
	}}, types.TypeList{Child: types.Str})

	bad(&parser.ListExpr{}, "cannot determine type from empty list")
	bad(&parser.ListExpr{Elements: []parser.Expr{
		&parser.IdentExpr{Name: "a"},
	}}, "variable 'a' was used before it was declared")
	bad(&parser.ListExpr{Elements: []parser.Expr{
		&parser.StringExpr{Val: "foo"},
		&parser.NumberExpr{Val: 456},
	}}, "element type Int is not compatible with type Str")
}

func TestCheckSubscriptExpr(t *testing.T) {
	s := scope.MakeGlobalScope()
	str := &parser.StringExpr{Tok: nop, Val: "foo"}
	index := &parser.NumberExpr{Tok: nop, Val: 0}
	expr := &parser.SubscriptExpr{ListLike: str, Index: index}
	typ := checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.TypeOptional{Child: types.Str})

	s = scope.MakeGlobalScope()
	list := &parser.ListExpr{Elements: []parser.Expr{
		&parser.NumberExpr{Val: 123},
		&parser.NumberExpr{Val: 456},
	}}
	expr = &parser.SubscriptExpr{ListLike: list, Index: index}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.TypeOptional{Child: types.Int})

	s = scope.MakeGlobalScope()
	str = &parser.StringExpr{Tok: nop, Val: "foo"}
	badRef := &parser.IdentExpr{Tok: nop, Name: "x"}
	expr = &parser.SubscriptExpr{ListLike: str, Index: badRef}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "variable 'x' was used before it was declared")
	expectBool(t, typ.IsError(), true)

	s = scope.MakeGlobalScope()
	str = &parser.StringExpr{Tok: nop, Val: "foo"}
	badIndex := &parser.StringExpr{Tok: nop, Val: "0"}
	expr = &parser.SubscriptExpr{ListLike: str, Index: badIndex}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "subscript operator does not support Str[Str]")
	expectBool(t, typ.IsError(), true)

	s = scope.MakeGlobalScope()
	str = &parser.StringExpr{Tok: nop, Val: "foo"}
	expr = &parser.SubscriptExpr{ListLike: str, Index: index}
	typ = checkSubscriptExpr(s, expr, make(binopsLUT))
	expectNthError(t, s, 0, "unknown infix operator '['")
	expectBool(t, typ.IsError(), true)
}

func TestCheckSelfExpr(t *testing.T) {
	prog, _ := parser.Parse("let f := fn(): Void { self(); };")
	s := Check(prog)
	expectNoErrors(t, s)

	prog, _ = parser.Parse("self();")
	s = Check(prog)
	expectNthError(t, s, 0, "self references must be inside a function")
}

func TestCheckIdentExpr(t *testing.T) {
	s := scope.MakeGlobalScope()
	s.NewVariable("x", types.Int)
	expr := &parser.IdentExpr{Tok: nop, Name: "x"}
	typ := checkIdentExpr(s, expr)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Int)

	s = scope.MakeGlobalScope()
	expr = &parser.IdentExpr{Tok: nop, Name: "x"}
	typ = checkIdentExpr(s, expr)
	expectNthError(t, s, 0, "variable 'x' was used before it was declared")
	expectBool(t, typ.IsError(), true)
}

func TestCheckNumberExpr(t *testing.T) {
	s := scope.MakeGlobalScope()
	expr := &parser.NumberExpr{Tok: nop, Val: 123}
	typ := checkNumberExpr(s, expr)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Int)
}

func TestCheckStringExpr(t *testing.T) {
	s := scope.MakeGlobalScope()
	expr := &parser.StringExpr{Tok: nop, Val: "abc"}
	typ := checkStringExpr(s, expr)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Str)
}

func TestCheckBooleanExpr(t *testing.T) {
	s := scope.MakeGlobalScope()
	expr := &parser.BooleanExpr{Tok: nop, Val: true}
	typ := checkBooleanExpr(s, expr)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Bool)
}

func TestConvertTypeSig(t *testing.T) {
	var note parser.TypeNote

	note = parser.TypeNoteVoid{Tok: nop}
	expectEquivalentType(t, ConvertTypeNote(note), types.TypeVoid{})

	note = parser.TypeNoteFunction{
		Params: parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
			parser.TypeNoteIdent{Tok: nop, Name: "Int"},
			parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
		}},
		Ret: parser.TypeNoteIdent{Tok: nop, Name: "Str"},
	}
	expectEquivalentType(t, ConvertTypeNote(note), types.TypeFunction{
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
	expectEquivalentType(t, ConvertTypeNote(note), types.TypeFunction{
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
	expectEquivalentType(t, ConvertTypeNote(note), types.TypeTuple{Children: []types.Type{
		types.TypeIdent{Name: "Int"},
		types.TypeIdent{Name: "Bool"},
	}})

	note = parser.TypeNoteList{Tok: nop, Child: parser.TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, ConvertTypeNote(note), types.TypeList{Child: types.TypeIdent{Name: "Int"}})

	note = parser.TypeNoteOptional{Tok: nop, Child: parser.TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, ConvertTypeNote(note), types.TypeOptional{Child: types.TypeIdent{Name: "Int"}})

	note = parser.TypeNoteIdent{Tok: nop, Name: "Int"}
	expectEquivalentType(t, ConvertTypeNote(note), types.TypeIdent{Name: "Int"})

	note = nil
	expectBool(t, ConvertTypeNote(note) == nil, true)
}

func expectNthError(t *testing.T, s scope.Scope, n int, msg string) {
	if len(s.GetErrors()) <= n {
		t.Fatalf("Expected at least %d errors", n+1)
	}

	expectAnError(t, s.GetErrors()[n], msg)
}

func expectLocalVariableType(t *testing.T, s scope.Scope, name string, exp types.Type) {
	if s.HasLocalVariable(name) {
		got := s.GetLocalVariableType(name)
		expectEquivalentType(t, got, exp)
	} else {
		t.Errorf("Expected local variable '%s', none found", name)
	}
}

func expectNoErrors(t *testing.T, s scope.Scope) {
	if s.HasErrors() {
		for i, err := range s.GetErrors() {
			t.Errorf("%d '%s'", i, err)
		}

		t.Fatalf("Expected no errors, found %d", len(s.GetErrors()))
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
