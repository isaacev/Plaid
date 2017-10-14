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
	s := Check(&parser.Program{})
	expectNoErrors(t, s)

	mod1 := &vm.Module{
		Name: "mod1",
		Exports: map[string]*vm.Export{
			"foo": &vm.Export{
				Type: types.Bool,
			},
		},
	}
	s = Check(&parser.Program{}, mod1)
	expectVariable(t, s, "foo", types.Bool)
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
	expectNthError(t, s, 0, "(1:4) condition must resolve to a boolean")
}

func TestCheckReturnStmt(t *testing.T) {
	prog, _ := parser.Parse("let a := fn (): Int { return \"abc\"; };")
	s := Check(prog)
	expectNthError(t, s, 0, "(1:30) expected to return 'Int', got 'Str'")

	prog, _ = parser.Parse("let a := fn (): Int { return x; };")
	s = Check(prog)
	expectNthError(t, s, 0, "(1:30) variable 'x' was used before it was declared")

	prog, _ = parser.Parse("let a := fn (): Int { return; };")
	s = Check(prog)
	expectNthError(t, s, 0, "(1:23) expected a return type of 'Int', got nothing")

	prog, _ = parser.Parse("let a := fn ():Void { return 123; };")
	s = Check(prog)
	expectNthError(t, s, 0, "(1:30) expected to return nothing, got 'Int'")

	prog, _ = parser.Parse("return;")
	s = Check(prog)
	expectNthError(t, s, 0, "(1:1) return statements must be inside a function")
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
	expectNthError(t, s, 0, "(1:10) variable 'add' was used before it was declared")
	expectBool(t, s.GetVariableType("a").IsError(), true)

	prog, _ = parser.Parse("let f := fn():Void{}; let a := f();")
	s = Check(prog)
	expectNthError(t, s, 0, "(1:32) cannot use void types in an expression")
	expectBool(t, s.GetVariableType("a").IsError(), true)

	prog, _ = parser.Parse("let a := -5;")
	s = Check(prog)
	expectNthError(t, s, 0, "(1:10) unknown expression type")
	expectBool(t, s.GetVariableType("a").IsError(), true)
}

func TestCheckFunctionExpr(t *testing.T) {
	prog, _ := parser.Parse("let f := fn (a: Int): Int { };")
	s := Check(prog)
	expectNoErrors(t, s)
	expectEquivalentType(t, s.GetVariableType("f"), types.Function{
		Params: types.Tuple{Children: []types.Type{types.Ident{Name: "Int"}}},
		Ret:    types.Ident{Name: "Int"},
	})
}

func TestCheckDispatchExpr(t *testing.T) {
	good := func(source string, name string, typ types.Type) {
		if prog, err := parser.Parse(source); err != nil {
			t.Fatal(err)
		} else {
			s := scope.MakeGlobalScope()
			s.NewVariable(name, typ)
			checkProgram(s, prog)
			expectNoErrors(t, s)
		}
	}

	bad := func(source string, name string, typ types.Type, errs ...string) {
		if prog, err := parser.Parse(source); err != nil {
			t.Fatal(err)
		} else {
			s := scope.MakeGlobalScope()
			s.NewVariable(name, typ)
			checkProgram(s, prog)
			for i, err := range errs {
				expectNthError(t, s, i, err)
			}
		}
	}

	good("add(2, 5);", "add", types.Function{
		Params: types.Tuple{Children: []types.Type{
			types.Ident{Name: "Int"},
			types.Ident{Name: "Int"},
		}},
		Ret: types.Ident{Name: "Int"},
	})

	bad("add(2, 5);", "add", types.Int, "(1:1) cannot call function on type 'Int'")
	bad("add(2);", "add", types.Function{
		Params: types.Tuple{Children: []types.Type{
			types.Ident{Name: "Int"},
			types.Ident{Name: "Int"},
		}},
		Ret: types.Ident{Name: "Int"},
	}, "(1:1) expected 2 arguments, got 1")
	bad("self();", "", nil, "(1:1) self references must be inside a function")
	bad(`add("2", "4");`, "add", types.Function{
		Params: types.Tuple{Children: []types.Type{
			types.Ident{Name: "Int"},
			types.Ident{Name: "Int"},
		}},
		Ret: types.Ident{Name: "Int"},
	}, "(1:5) expected 'Int', got 'Str'", "(1:10) expected 'Int', got 'Str'")
}

func TestCheckAssignExpr(t *testing.T) {
	good := func(source string, name string, typ types.Type) {
		if prog, err := parser.Parse(source); err != nil {
			t.Fatal(err)
		} else {
			s := scope.MakeGlobalScope()
			s.NewVariable(name, typ)
			checkProgram(s, prog)
			expectNoErrors(t, s)
		}
	}

	bad := func(source string, name string, typ types.Type, exp string) {
		if prog, err := parser.Parse(source); err != nil {
			t.Fatal(err)
		} else {
			s := scope.MakeGlobalScope()
			s.NewVariable(name, typ)
			checkProgram(s, prog)
			expectNthError(t, s, 0, exp)
		}
	}

	good("a := 456;", "a", types.Int)
	good("b := \"456\";", "b", types.Str)
	good("c := true;", "c", types.Bool)

	bad("a := 456;", "a", types.Str, "(1:6) 'Str' cannot be assigned type 'Int'")
	bad(`a := "a" + 45;`, "a", types.Str, "(1:10) operator '+' does not support Str and Int")
	bad("a := 123;", "b", types.Str, "(1:1) 'a' cannot be assigned before it is declared")
}

func TestCheckBinaryExpr(t *testing.T) {
	good := func(left types.Type, oper string, right types.Type, exp types.Type) {
		source := "let c := a " + oper + " b;"
		if prog, err := parser.Parse(source); err != nil {
			t.Fatal(err)
		} else {
			s := scope.MakeGlobalScope()
			s.NewVariable("a", left)
			s.NewVariable("b", right)
			checkProgram(s, prog)
			expectNoErrors(t, s)
			expectEquivalentType(t, s.GetVariableType("c"), exp)
		}
	}

	bad := func(source string, s scope.Scope, errs ...string) {
		if prog, err := parser.Parse(source); err != nil {
			t.Fatal(err)
		} else {
			checkProgram(s, prog)
			for n, err := range errs {
				expectNthError(t, s, n, err)
			}
			expectEquivalentType(t, s.GetVariableType("c"), types.Error{})
		}
	}

	good(types.Int, "+", types.Int, types.Int)
	good(types.Int, "-", types.Int, types.Int)

	s := scope.MakeGlobalScope()
	s.NewVariable("b", types.Int)
	bad("let c := a + b;", s,
		"(1:10) variable 'a' was used before it was declared")

	s = scope.MakeGlobalScope()
	bad("let c := a + b;", s,
		"(1:10) variable 'a' was used before it was declared",
		"(1:14) variable 'b' was used before it was declared")

	s = scope.MakeGlobalScope()
	s.NewVariable("a", types.Int)
	s.NewVariable("b", types.Int)
	oper := lexer.Token{Loc: lexer.Loc{Line: 10, Col: 4}}
	leftExpr := &parser.IdentExpr{Name: "a"}
	rightExpr := &parser.IdentExpr{Name: "b"}
	expr := &parser.BinaryExpr{Tok: oper, Oper: "@", Left: leftExpr, Right: rightExpr}
	typ := checkBinaryExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "(10:4) unknown infix operator '@'")
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
		expectEquivalentType(t, got, types.Error{})
	}

	good(&parser.ListExpr{Elements: []parser.Expr{
		&parser.StringExpr{Val: "foo"},
	}}, types.List{Child: types.Str})

	start := makeTok(5, 4)
	first := makeTok(7, 12)
	second := makeTok(3, 9)
	bad(&parser.ListExpr{Tok: start}, "(5:4) cannot determine type from empty list")
	bad(&parser.ListExpr{Tok: start, Elements: []parser.Expr{
		&parser.IdentExpr{Tok: first, Name: "a"},
	}}, "(7:12) variable 'a' was used before it was declared")
	bad(&parser.ListExpr{Tok: start, Elements: []parser.Expr{
		&parser.StringExpr{Tok: first, Val: "foo"},
		&parser.NumberExpr{Tok: second, Val: 456},
	}}, "(3:9) element type Int is not compatible with type Str")
}

func TestCheckSubscriptExpr(t *testing.T) {
	s := scope.MakeGlobalScope()
	str := &parser.StringExpr{Tok: nop, Val: "foo"}
	index := &parser.NumberExpr{Tok: nop, Val: 0}
	expr := &parser.SubscriptExpr{ListLike: str, Index: index}
	typ := checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Optional{Child: types.Str})

	s = scope.MakeGlobalScope()
	list := &parser.ListExpr{Elements: []parser.Expr{
		&parser.NumberExpr{Val: 123},
		&parser.NumberExpr{Val: 456},
	}}
	expr = &parser.SubscriptExpr{ListLike: list, Index: index}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Optional{Child: types.Int})

	s = scope.MakeGlobalScope()
	str = &parser.StringExpr{Tok: nop, Val: "foo"}
	badRef := &parser.IdentExpr{Tok: makeTok(5, 2), Name: "x"}
	expr = &parser.SubscriptExpr{ListLike: str, Index: badRef}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "(5:2) variable 'x' was used before it was declared")
	expectBool(t, typ.IsError(), true)

	s = scope.MakeGlobalScope()
	str = &parser.StringExpr{Tok: nop, Val: "foo"}
	badIndex := &parser.StringExpr{Tok: makeTok(2, 9), Val: "0"}
	expr = &parser.SubscriptExpr{ListLike: str, Index: badIndex}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "(2:9) subscript operator does not support Str[Str]")
	expectBool(t, typ.IsError(), true)

	s = scope.MakeGlobalScope()
	str = &parser.StringExpr{Tok: makeTok(4, 2), Val: "foo"}
	expr = &parser.SubscriptExpr{ListLike: str, Index: index}
	typ = checkSubscriptExpr(s, expr, make(binopsLUT))
	expectNthError(t, s, 0, "(4:2) unknown infix operator '['")
	expectBool(t, typ.IsError(), true)
}

func TestCheckSelfExpr(t *testing.T) {
	prog, _ := parser.Parse("let f := fn(): Void { self(); };")
	s := Check(prog)
	expectNoErrors(t, s)

	prog, _ = parser.Parse("self();")
	s = Check(prog)
	expectNthError(t, s, 0, "(1:1) self references must be inside a function")
}

func TestCheckIdentExpr(t *testing.T) {
	s := scope.MakeGlobalScope()
	s.NewVariable("x", types.Int)
	expr := &parser.IdentExpr{Tok: nop, Name: "x"}
	typ := checkIdentExpr(s, expr)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Int)

	s = scope.MakeGlobalScope()
	expr = &parser.IdentExpr{Tok: makeTok(10, 13), Name: "x"}
	typ = checkIdentExpr(s, expr)
	expectNthError(t, s, 0, "(10:13) variable 'x' was used before it was declared")
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
	expectEquivalentType(t, ConvertTypeNote(note), types.Void{})

	note = parser.TypeNoteFunction{
		Params: parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
			parser.TypeNoteIdent{Tok: nop, Name: "Int"},
			parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
		}},
		Ret: parser.TypeNoteIdent{Tok: nop, Name: "Str"},
	}
	expectEquivalentType(t, ConvertTypeNote(note), types.Function{
		Params: types.Tuple{Children: []types.Type{
			types.Ident{Name: "Int"},
			types.Ident{Name: "Bool"},
		}},
		Ret: types.Ident{Name: "Str"},
	})

	note = parser.TypeNoteFunction{
		Params: parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
			parser.TypeNoteIdent{Tok: nop, Name: "Int"},
			parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
		}},
		Ret: parser.TypeNoteVoid{},
	}
	expectEquivalentType(t, ConvertTypeNote(note), types.Function{
		Params: types.Tuple{Children: []types.Type{
			types.Ident{Name: "Int"},
			types.Ident{Name: "Bool"},
		}},
		Ret: types.Void{},
	})

	note = parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
		parser.TypeNoteIdent{Tok: nop, Name: "Int"},
		parser.TypeNoteIdent{Tok: nop, Name: "Bool"},
	}}
	expectEquivalentType(t, ConvertTypeNote(note), types.Tuple{Children: []types.Type{
		types.Ident{Name: "Int"},
		types.Ident{Name: "Bool"},
	}})

	note = parser.TypeNoteList{Tok: nop, Child: parser.TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, ConvertTypeNote(note), types.List{Child: types.Ident{Name: "Int"}})

	note = parser.TypeNoteOptional{Tok: nop, Child: parser.TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, ConvertTypeNote(note), types.Optional{Child: types.Ident{Name: "Int"}})

	note = parser.TypeNoteIdent{Tok: nop, Name: "Int"}
	expectEquivalentType(t, ConvertTypeNote(note), types.Ident{Name: "Int"})

	note = nil
	expectBool(t, ConvertTypeNote(note) == nil, true)
}

func makeTok(line int, col int) lexer.Token {
	return lexer.Token{Loc: lexer.Loc{Line: line, Col: col}}
}

func expectNthError(t *testing.T, s scope.Scope, n int, msg string) {
	if len(s.GetErrors()) <= n {
		t.Fatalf("Expected at least %d errors", n+1)
	}

	expectAnError(t, s.GetErrors()[n], msg)
}

func expectVariable(t *testing.T, s scope.Scope, name string, exp types.Type) {
	if s.HasVariable(name) {
		got := s.GetVariableType(name)
		expectEquivalentType(t, got, exp)
	} else {
		t.Errorf("Expected variable '%s', none found", name)
	}
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
