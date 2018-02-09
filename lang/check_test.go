package lang

import (
	"plaid/lang/types"
	"testing"
)

func TestCheckProgram(t *testing.T) {
	good := func(src string) {
		ast, _ := Parse("", src)
		s := makeScope(nil)
		checkProgram(s, ast)
		expectNoXScopeErrors(t, s)
	}

	good("let a := 123;")
}

func TestCheckStmt(t *testing.T) {
	good := func(src string) {
		ast, _ := Parse("", src)
		s := makeScope(nil)
		checkStmt(s, ast.Stmts[0])
		expectNoXScopeErrors(t, s)
	}

	good("let a := 123;")
}

func TestCheckPubStmt(t *testing.T) {
	good := func(src string) {
		ast, _ := Parse("", src)
		s := makeScope(nil)
		s.Module = &ModuleVirtual{}
		checkProgram(s, ast)
		expectNoXScopeErrors(t, s)
	}

	bad := func(src string, msg string) {
		ast, _ := Parse("", src)
		s1 := makeScope(nil)
		s1.Module = &ModuleVirtual{}
		s2 := makeScope(s1)
		checkProgram(s2, ast)
		expectNthXScopeError(t, s2, 0, msg)
	}

	good("pub let x := 100;")
	bad("pub let x := 100;", "(1:1) pub statement must be a top-level statement")
}

func TestCheckIfStmt(t *testing.T) {
	goodProgram(t, "if true {};")
	badProgram(t, "if 123 {};", "(1:4) condition must resolve to a boolean")
}

func TestCheckReturnStmt(t *testing.T) {
	badProgram(t,
		"let a := fn (): Int { return \"abc\"; };",
		"(1:30) expected to return 'Int', got 'Str'")

	badProgram(t,
		"let a := fn (): Int { return x; };",
		"(1:30) variable 'x' was used before it was declared")

	badProgram(t,
		"let a := fn (): Int { return; };",
		"(1:23) expected a return type of 'Int', got nothing")

	badProgram(t,
		"let a := fn ():Void { return 123; };",
		"(1:30) expected to return nothing, got 'Int'")

	prog := &RootNode{Stmts: []Stmt{
		&ReturnStmt{Tok: makeTok(1, 1)},
	}}
	s := checkProgram(makeScope(nil), prog)
	expectNthXScopeError(t, s, 0, "(1:1) return statements must be inside a function")
}

func TestCheckExpr(t *testing.T) {
	prog, _ := Parse("", "let a := 2 + 1;")
	s := checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, s.Lookup("a"), types.BuiltinInt)

	prog, _ = Parse("", "let a := 1;")
	s = checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, s.Lookup("a"), types.BuiltinInt)

	prog, _ = Parse("", "let a := \"abc\";")
	s = checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, s.Lookup("a"), types.BuiltinStr)

	prog, _ = Parse("", "let a := fn () {};")
	s = checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)

	prog, _ = Parse("", "let a := true;")
	s = checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)

	prog, _ = Parse("", "let a := false;")
	s = checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)

	prog, _ = Parse("", "let a := [1, 2, 3];")
	s = checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)

	prog, _ = Parse("", "let a := \"abc\"[0];")
	s = checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)

	prog, _ = Parse("", "let a := add(2, 2);")
	s = checkProgram(makeScope(nil), prog)
	expectNthXScopeError(t, s, 0, "(1:10) variable 'add' was used before it was declared")
	expectBool(t, s.Lookup("a").IsError(), true)

	prog, _ = Parse("", "let f := fn():Void{}; let a := f();")
	s = checkProgram(makeScope(nil), prog)
	expectNthXScopeError(t, s, 0, "(1:32) cannot use void types in an expression")
	expectBool(t, s.Lookup("a").IsError(), true)

	prog, _ = Parse("", "let a := -5;")
	s = checkProgram(makeScope(nil), prog)
	expectNthXScopeError(t, s, 0, "(1:10) unknown expression type")
	expectBool(t, s.Lookup("a").IsError(), true)
}

func TestCheckFunctionExpr(t *testing.T) {
	prog, _ := Parse("", "let f := fn (a: Int): Int { };")
	s := checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, s.Lookup("f"), types.Function{
		Params: types.Tuple{Children: []types.Type{types.Ident{Name: "Int"}}},
		Ret:    types.Ident{Name: "Int"},
	})
}

func TestCheckDispatchExpr(t *testing.T) {
	good := func(source string, name string, typ types.Type) {
		t.Helper()
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeScope(nil)
			s.AddLocal(name, typ)
			checkProgram(s, prog)
			expectNoXScopeErrors(t, s)
		}
	}

	bad := func(source string, name string, typ types.Type, errs ...string) {
		t.Helper()
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeScope(nil)
			s.AddLocal(name, typ)
			checkProgram(s, prog)
			for i, err := range errs {
				expectNthXScopeError(t, s, i, err)
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

	bad("add(2, 5);", "add", types.BuiltinInt, "(1:1) cannot call function on type 'Int'")
	bad("add(2);", "add", types.Function{
		Params: types.Tuple{Children: []types.Type{
			types.Ident{Name: "Int"},
			types.Ident{Name: "Int"},
		}},
		Ret: types.Ident{Name: "Int"},
	}, "(1:1) expected 2 arguments, got 1")
	bad("self();", "", nil, "(1:1) self references must be inside a function")
	bad("add(5, x);", "add", types.Function{
		Params: types.Tuple{Children: []types.Type{
			types.Ident{Name: "Int"},
			types.Ident{Name: "Int"},
		}},
		Ret: types.Ident{Name: "Int"},
	}, "(1:8) variable 'x' was used before it was declared")
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
		t.Helper()
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeScope(nil)
			s.AddLocal(name, typ)
			checkProgram(s, prog)
			expectNoXScopeErrors(t, s)
		}
	}

	bad := func(source string, name string, typ types.Type, exp string) {
		t.Helper()
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeScope(nil)
			s.AddLocal(name, typ)
			checkProgram(s, prog)
			expectNthXScopeError(t, s, 0, exp)
		}
	}

	good("a := 456;", "a", types.BuiltinInt)
	good("b := \"456\";", "b", types.BuiltinStr)
	good("c := true;", "c", types.BuiltinBool)

	bad("a := 456;", "a", types.BuiltinStr, "(1:6) 'Str' cannot be assigned type 'Int'")
	bad(`a := "a" + 45;`, "a", types.BuiltinStr, "(1:10) operator '+' does not support Str and Int")
	bad("a := 123;", "b", types.BuiltinStr, "(1:1) 'a' cannot be assigned before it is declared")
}

func TestCheckBinaryExpr(t *testing.T) {
	good := func(left types.Type, oper string, right types.Type, exp types.Type) {
		t.Helper()
		source := "let c := a " + oper + " b;"
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeScope(nil)
			s.AddLocal("a", left)
			s.AddLocal("b", right)
			checkProgram(s, prog)
			expectNoXScopeErrors(t, s)
			expectEquivalentType(t, s.Lookup("c"), exp)
		}
	}

	bad := func(source string, s *Scope, errs ...string) {
		t.Helper()
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			checkProgram(s, prog)
			for n, err := range errs {
				expectNthXScopeError(t, s, n, err)
			}
			expectEquivalentType(t, s.Lookup("c"), types.Error{})
		}
	}

	good(types.BuiltinInt, "+", types.BuiltinInt, types.BuiltinInt)
	good(types.BuiltinInt, "-", types.BuiltinInt, types.BuiltinInt)

	s := makeScope(nil)
	s.AddLocal("b", types.BuiltinInt)
	bad("let c := a + b;", s,
		"(1:10) variable 'a' was used before it was declared")

	s = makeScope(nil)
	bad("let c := a + b;", s,
		"(1:10) variable 'a' was used before it was declared",
		"(1:14) variable 'b' was used before it was declared")

	s = makeScope(nil)
	s.AddLocal("a", types.BuiltinInt)
	s.AddLocal("b", types.BuiltinInt)
	oper := token{Loc: Loc{Line: 10, Col: 4}}
	leftExpr := &IdentExpr{Name: "a"}
	rightExpr := &IdentExpr{Name: "b"}
	expr := &BinaryExpr{Tok: oper, Oper: "@", Left: leftExpr, Right: rightExpr}
	typ := checkBinaryExpr(s, expr, defaultBinopsLUT)
	expectNthXScopeError(t, s, 0, "(10:4) unknown infix operator '@'")
	expectBool(t, typ.IsError(), true)
}

func TestCheckListExpr(t *testing.T) {
	good := func(expr *ListExpr, exp types.Type) {
		t.Helper()
		s := makeScope(nil)
		got := checkListExpr(s, expr)
		expectNoXScopeErrors(t, s)
		expectEquivalentType(t, got, exp)
	}

	bad := func(expr *ListExpr, exp string) {
		t.Helper()
		s := makeScope(nil)
		got := checkListExpr(s, expr)
		expectNthXScopeError(t, s, 0, exp)
		expectEquivalentType(t, got, types.Error{})
	}

	good(&ListExpr{Elements: []Expr{
		&StringExpr{Val: "foo"},
	}}, types.List{Child: types.BuiltinStr})

	start := makeTok(5, 4)
	first := makeTok(7, 12)
	second := makeTok(3, 9)
	bad(&ListExpr{Tok: start}, "(5:4) cannot determine type from empty list")
	bad(&ListExpr{Tok: start, Elements: []Expr{
		&IdentExpr{Tok: first, Name: "a"},
	}}, "(7:12) variable 'a' was used before it was declared")
	bad(&ListExpr{Tok: start, Elements: []Expr{
		&StringExpr{Tok: first, Val: "foo"},
		&NumberExpr{Tok: second, Val: 456},
	}}, "(3:9) element type Int is not compatible with type Str")
}

func TestCheckSubscriptExpr(t *testing.T) {
	s := makeScope(nil)
	str := &StringExpr{Tok: nop, Val: "foo"}
	index := &NumberExpr{Tok: nop, Val: 0}
	expr := &SubscriptExpr{ListLike: str, Index: index}
	typ := checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, typ, types.Optional{Child: types.BuiltinStr})

	s = makeScope(nil)
	list := &ListExpr{Elements: []Expr{
		&NumberExpr{Val: 123},
		&NumberExpr{Val: 456},
	}}
	expr = &SubscriptExpr{ListLike: list, Index: index}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, typ, types.Optional{Child: types.BuiltinInt})

	s = makeScope(nil)
	str = &StringExpr{Tok: nop, Val: "foo"}
	badRef := &IdentExpr{Tok: makeTok(5, 2), Name: "x"}
	expr = &SubscriptExpr{ListLike: str, Index: badRef}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNthXScopeError(t, s, 0, "(5:2) variable 'x' was used before it was declared")
	expectBool(t, typ.IsError(), true)

	s = makeScope(nil)
	str = &StringExpr{Tok: nop, Val: "foo"}
	badIndex := &StringExpr{Tok: makeTok(2, 9), Val: "0"}
	expr = &SubscriptExpr{ListLike: str, Index: badIndex}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNthXScopeError(t, s, 0, "(2:9) subscript operator does not support Str[Str]")
	expectBool(t, typ.IsError(), true)

	s = makeScope(nil)
	str = &StringExpr{Tok: makeTok(4, 2), Val: "foo"}
	expr = &SubscriptExpr{ListLike: str, Index: index}
	typ = checkSubscriptExpr(s, expr, make(binopsLUT))
	expectNthXScopeError(t, s, 0, "(4:2) unknown infix operator '['")
	expectBool(t, typ.IsError(), true)
}

func TestCheckAccessExpr(t *testing.T) {
	/*
		good := func(src string) {
			t.Helper()
			ast, _ := Parse("", src)
			s := checkProgram(makeXScope(nil), ast)
			expectNoXScopeErrors(t, s)
		}

		good("let a := [1,2]; let b := a.length();")
	*/
}

func TestCheckSelfExpr(t *testing.T) {
	prog, _ := Parse("", "let f := fn(): Void { self(); };")
	s := checkProgram(makeScope(nil), prog)
	expectNoXScopeErrors(t, s)

	prog, _ = Parse("", "self();")
	s = checkProgram(makeScope(nil), prog)
	expectNthXScopeError(t, s, 0, "(1:1) self references must be inside a function")
}

func TestCheckIdentExpr(t *testing.T) {
	s := makeScope(nil)
	s.AddLocal("x", types.BuiltinInt)
	expr := &IdentExpr{Tok: nop, Name: "x"}
	typ := checkIdentExpr(s, expr)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, typ, types.BuiltinInt)

	s = makeScope(nil)
	expr = &IdentExpr{Tok: makeTok(10, 13), Name: "x"}
	typ = checkIdentExpr(s, expr)
	expectNthXScopeError(t, s, 0, "(10:13) variable 'x' was used before it was declared")
	expectBool(t, typ.IsError(), true)
}

func TestCheckNumberExpr(t *testing.T) {
	s := makeScope(nil)
	expr := &NumberExpr{Tok: nop, Val: 123}
	typ := checkNumberExpr(s, expr)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, typ, types.BuiltinInt)
}

func TestCheckStringExpr(t *testing.T) {
	s := makeScope(nil)
	expr := &StringExpr{Tok: nop, Val: "abc"}
	typ := checkStringExpr(s, expr)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, typ, types.BuiltinStr)
}

func TestCheckBooleanExpr(t *testing.T) {
	s := makeScope(nil)
	expr := &BooleanExpr{Tok: nop, Val: true}
	typ := checkBooleanExpr(s, expr)
	expectNoXScopeErrors(t, s)
	expectEquivalentType(t, typ, types.BuiltinBool)
}

func TestConvertTypeSig(t *testing.T) {
	var note TypeNote

	note = TypeNoteVoid{Tok: nop}
	expectEquivalentType(t, convertTypeNote(note), types.Void{})

	note = TypeNoteFunction{
		Params: TypeNoteTuple{Tok: nop, Elems: []TypeNote{
			TypeNoteIdent{Tok: nop, Name: "Int"},
			TypeNoteIdent{Tok: nop, Name: "Bool"},
		}},
		Ret: TypeNoteIdent{Tok: nop, Name: "Str"},
	}
	expectEquivalentType(t, convertTypeNote(note), types.Function{
		Params: types.Tuple{Children: []types.Type{
			types.Ident{Name: "Int"},
			types.Ident{Name: "Bool"},
		}},
		Ret: types.Ident{Name: "Str"},
	})

	note = TypeNoteFunction{
		Params: TypeNoteTuple{Tok: nop, Elems: []TypeNote{
			TypeNoteIdent{Tok: nop, Name: "Int"},
			TypeNoteIdent{Tok: nop, Name: "Bool"},
		}},
		Ret: TypeNoteVoid{},
	}
	expectEquivalentType(t, convertTypeNote(note), types.Function{
		Params: types.Tuple{Children: []types.Type{
			types.Ident{Name: "Int"},
			types.Ident{Name: "Bool"},
		}},
		Ret: types.Void{},
	})

	note = TypeNoteTuple{Tok: nop, Elems: []TypeNote{
		TypeNoteIdent{Tok: nop, Name: "Int"},
		TypeNoteIdent{Tok: nop, Name: "Bool"},
	}}
	expectEquivalentType(t, convertTypeNote(note), types.Tuple{Children: []types.Type{
		types.Ident{Name: "Int"},
		types.Ident{Name: "Bool"},
	}})

	note = TypeNoteList{Tok: nop, Child: TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, convertTypeNote(note), types.List{Child: types.Ident{Name: "Int"}})

	note = TypeNoteOptional{Tok: nop, Child: TypeNoteIdent{Tok: nop, Name: "Int"}}
	expectEquivalentType(t, convertTypeNote(note), types.Optional{Child: types.Ident{Name: "Int"}})

	note = TypeNoteIdent{Tok: nop, Name: "Int"}
	expectEquivalentType(t, convertTypeNote(note), types.Ident{Name: "Int"})

	note = nil
	expectBool(t, convertTypeNote(note) == nil, true)
}

func TestConvertTypeNote(t *testing.T) {
	var note TypeNote
	var blob TypeNote = TypeNoteIdent{Tok: nop, Name: "Blob"}
	var blub TypeNote = TypeNoteIdent{Tok: nop, Name: "Blub"}
	var blah TypeNote = TypeNoteIdent{Tok: nop, Name: "Blah"}

	note = TypeNoteVoid{}
	expectConversion(t, note, "Void")

	note = TypeNoteAny{}
	expectConversion(t, note, "Any")

	note = TypeNoteFunction{
		Params: TypeNoteTuple{Tok: nop, Elems: []TypeNote{}},
		Ret:    TypeNoteVoid{},
	}
	expectConversion(t, note, "() => Void")

	note = TypeNoteFunction{
		Params: TypeNoteTuple{Tok: nop, Elems: []TypeNote{
			blob,
			blub,
		}},
		Ret: blah,
	}
	expectConversion(t, note, "(Blob Blub) => Blah")

	note = TypeNoteList{Tok: nop, Child: blob}
	expectConversion(t, note, "[Blob]")

	note = TypeNoteOptional{Tok: nop, Child: blah}
	expectConversion(t, note, "Blah?")

	note = nil
	got := convertTypeNote(note)
	if got != nil {
		t.Errorf("Expected '%v', got '%v'", nil, got)
	}
}

func goodProgram(t *testing.T, src string) {
	t.Helper()
	ast, _ := Parse("", src)
	s := makeScope(nil)
	checkProgram(s, ast)
	expectNoXScopeErrors(t, s)
}

func badProgram(t *testing.T, src string, msg string) {
	t.Helper()
	ast, _ := Parse("", src)
	s := makeScope(nil)
	checkProgram(s, ast)
	expectNthXScopeError(t, s, 0, msg)
}

func expectConversion(t *testing.T, note TypeNote, exp string) {
	t.Helper()
	got := convertTypeNote(note)
	if got.String() != exp {
		t.Errorf("Expected '%s', got '%v'", exp, got)
	}
}

func expectNoXScopeErrors(t *testing.T, s *Scope) {
	t.Helper()
	if len(s.AllErrors()) > 0 {
		for i, err := range s.AllErrors() {
			t.Errorf("%d '%s'", i, err)
		}

		t.Fatalf("Expected no errors, found %d", len(s.AllErrors()))
	}
}

func expectNthXScopeError(t *testing.T, scope *Scope, n int, msg string) {
	t.Helper()
	if len(scope.AllErrors()) <= n {
		t.Fatalf("Expected at least %d errors", n+1)
	}
	expectAnError(t, scope.AllErrors()[n], msg)
}

func makeTok(line int, col int) token {
	return token{Loc: Loc{Line: line, Col: col}}
}
