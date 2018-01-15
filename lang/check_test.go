package lang

import (
	"plaid/lang/types"
	"testing"
)

func TestCheckMain(t *testing.T) {
	scope := checkModule(&VirtualModule{ast: &RootNode{}})
	expectNoErrors(t, scope)

	// mod1 := &vm.Module{
	// 	Name: "mod1",
	// 	Exports: map[string]*vm.Export{
	// 		"foo": &vm.Export{
	// 			types.Type: Bool,
	// 		},
	// 	},
	// }
	// s = Check(&Module{AST: &Program{}}, MakeGlobalScopeFromModule(mod1))
	// expectVariable(t, s, "foo", Bool)
	// expectNoErrors(t, s)
}

func TestCheckProgram(t *testing.T) {
	prog, _ := Parse("", "let a := 123;")
	s := makeGlobalScope()
	checkProgram(s, prog)
	expectNoErrors(t, s)
}

func TestCheckStmt(t *testing.T) {
	prog, _ := Parse("", "let a := 123;")
	s := makeGlobalScope()
	checkStmt(s, prog.Stmts[0])
	expectNoErrors(t, s)
}

func TestCheckPubStmt(t *testing.T) {
	prog, _ := Parse("", "pub let x := 100;")
	s := checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)

	prog, _ = Parse("", "pub let x := 100;")
	s = makeGlobalScope()
	checkPubStmt(makeLocalScope(s, types.Function{}), prog.Stmts[0].(*PubStmt))
	expectNthError(t, s, 0, "(1:1) pub statement must be a top-level statement")
}

func TestCheckIfStmt(t *testing.T) {
	prog, _ := Parse("", "if true {};")
	s := checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)

	prog, _ = Parse("", "if 123 {};")
	s = checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:4) condition must resolve to a boolean")
}

func TestCheckReturnStmt(t *testing.T) {
	prog, _ := Parse("", "let a := fn (): Int { return \"abc\"; };")
	s := checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:30) expected to return 'Int', got 'Str'")

	prog, _ = Parse("", "let a := fn (): Int { return x; };")
	s = checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:30) variable 'x' was used before it was declared")

	prog, _ = Parse("", "let a := fn (): Int { return; };")
	s = checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:23) expected a return type of 'Int', got nothing")

	prog, _ = Parse("", "let a := fn ():Void { return 123; };")
	s = checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:30) expected to return nothing, got 'Int'")

	prog = &RootNode{Stmts: []Stmt{
		&ReturnStmt{Tok: makeTok(1, 1)},
	}}
	s = checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:1) return statements must be inside a function")
}

func TestCheckExpr(t *testing.T) {
	prog, _ := Parse("", "let a := 2 + 1;")
	s := checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)
	expectEquivalentType(t, s.GetVariableType("a"), types.BuiltinInt)

	prog, _ = Parse("", "let a := 1;")
	s = checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)
	expectEquivalentType(t, s.GetVariableType("a"), types.BuiltinInt)

	prog, _ = Parse("", "let a := \"abc\";")
	s = checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)
	expectEquivalentType(t, s.GetVariableType("a"), types.BuiltinStr)

	prog, _ = Parse("", "let a := fn () {};")
	s = checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)

	prog, _ = Parse("", "let a := true;")
	s = checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)

	prog, _ = Parse("", "let a := false;")
	s = checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)

	prog, _ = Parse("", "let a := [1, 2, 3];")
	s = checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)

	prog, _ = Parse("", "let a := \"abc\"[0];")
	s = checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)

	prog, _ = Parse("", "let a := add(2, 2);")
	s = checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:10) variable 'add' was used before it was declared")
	expectBool(t, s.GetVariableType("a").IsError(), true)

	prog, _ = Parse("", "let f := fn():Void{}; let a := f();")
	s = checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:32) cannot use void types in an expression")
	expectBool(t, s.GetVariableType("a").IsError(), true)

	prog, _ = Parse("", "let a := -5;")
	s = checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:10) unknown expression type")
	expectBool(t, s.GetVariableType("a").IsError(), true)
}

func TestCheckFunctionExpr(t *testing.T) {
	prog, _ := Parse("", "let f := fn (a: Int): Int { };")
	s := checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)
	expectEquivalentType(t, s.GetVariableType("f"), types.Function{
		Params: types.Tuple{Children: []types.Type{types.Ident{Name: "Int"}}},
		Ret:    types.Ident{Name: "Int"},
	})
}

func TestCheckDispatchExpr(t *testing.T) {
	good := func(source string, name string, typ types.Type) {
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeGlobalScope()
			s.newVariable(name, typ)
			checkProgram(s, prog)
			expectNoErrors(t, s)
		}
	}

	bad := func(source string, name string, typ types.Type, errs ...string) {
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeGlobalScope()
			s.newVariable(name, typ)
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
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeGlobalScope()
			s.newVariable(name, typ)
			checkProgram(s, prog)
			expectNoErrors(t, s)
		}
	}

	bad := func(source string, name string, typ types.Type, exp string) {
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeGlobalScope()
			s.newVariable(name, typ)
			checkProgram(s, prog)
			expectNthError(t, s, 0, exp)
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
		source := "let c := a " + oper + " b;"
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			s := makeGlobalScope()
			s.newVariable("a", left)
			s.newVariable("b", right)
			checkProgram(s, prog)
			expectNoErrors(t, s)
			expectEquivalentType(t, s.GetVariableType("c"), exp)
		}
	}

	bad := func(source string, s *GlobalScope, errs ...string) {
		if prog, err := Parse("", source); err != nil {
			t.Fatal(err)
		} else {
			checkProgram(s, prog)
			for n, err := range errs {
				expectNthError(t, s, n, err)
			}
			expectEquivalentType(t, s.GetVariableType("c"), types.Error{})
		}
	}

	good(types.BuiltinInt, "+", types.BuiltinInt, types.BuiltinInt)
	good(types.BuiltinInt, "-", types.BuiltinInt, types.BuiltinInt)

	s := makeGlobalScope()
	s.newVariable("b", types.BuiltinInt)
	bad("let c := a + b;", s,
		"(1:10) variable 'a' was used before it was declared")

	s = makeGlobalScope()
	bad("let c := a + b;", s,
		"(1:10) variable 'a' was used before it was declared",
		"(1:14) variable 'b' was used before it was declared")

	s = makeGlobalScope()
	s.newVariable("a", types.BuiltinInt)
	s.newVariable("b", types.BuiltinInt)
	oper := token{Loc: Loc{Line: 10, Col: 4}}
	leftExpr := &IdentExpr{Name: "a"}
	rightExpr := &IdentExpr{Name: "b"}
	expr := &BinaryExpr{Tok: oper, Oper: "@", Left: leftExpr, Right: rightExpr}
	typ := checkBinaryExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "(10:4) unknown infix operator '@'")
	expectBool(t, typ.IsError(), true)
}

func TestCheckListExpr(t *testing.T) {
	good := func(expr *ListExpr, exp types.Type) {
		s := makeGlobalScope()
		got := checkListExpr(s, expr)
		expectNoErrors(t, s)
		expectEquivalentType(t, got, exp)
	}

	bad := func(expr *ListExpr, exp string) {
		s := makeGlobalScope()
		got := checkListExpr(s, expr)
		expectNthError(t, s, 0, exp)
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
	s := makeGlobalScope()
	str := &StringExpr{Tok: nop, Val: "foo"}
	index := &NumberExpr{Tok: nop, Val: 0}
	expr := &SubscriptExpr{ListLike: str, Index: index}
	typ := checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Optional{Child: types.BuiltinStr})

	s = makeGlobalScope()
	list := &ListExpr{Elements: []Expr{
		&NumberExpr{Val: 123},
		&NumberExpr{Val: 456},
	}}
	expr = &SubscriptExpr{ListLike: list, Index: index}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.Optional{Child: types.BuiltinInt})

	s = makeGlobalScope()
	str = &StringExpr{Tok: nop, Val: "foo"}
	badRef := &IdentExpr{Tok: makeTok(5, 2), Name: "x"}
	expr = &SubscriptExpr{ListLike: str, Index: badRef}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "(5:2) variable 'x' was used before it was declared")
	expectBool(t, typ.IsError(), true)

	s = makeGlobalScope()
	str = &StringExpr{Tok: nop, Val: "foo"}
	badIndex := &StringExpr{Tok: makeTok(2, 9), Val: "0"}
	expr = &SubscriptExpr{ListLike: str, Index: badIndex}
	typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	expectNthError(t, s, 0, "(2:9) subscript operator does not support Str[Str]")
	expectBool(t, typ.IsError(), true)

	s = makeGlobalScope()
	str = &StringExpr{Tok: makeTok(4, 2), Val: "foo"}
	expr = &SubscriptExpr{ListLike: str, Index: index}
	typ = checkSubscriptExpr(s, expr, make(binopsLUT))
	expectNthError(t, s, 0, "(4:2) unknown infix operator '['")
	expectBool(t, typ.IsError(), true)
}

func TestCheckSelfExpr(t *testing.T) {
	prog, _ := Parse("", "let f := fn(): Void { self(); };")
	s := checkProgram(makeGlobalScope(), prog)
	expectNoErrors(t, s)

	prog, _ = Parse("", "self();")
	s = checkProgram(makeGlobalScope(), prog)
	expectNthError(t, s, 0, "(1:1) self references must be inside a function")
}

func TestCheckIdentExpr(t *testing.T) {
	s := makeGlobalScope()
	s.newVariable("x", types.BuiltinInt)
	expr := &IdentExpr{Tok: nop, Name: "x"}
	typ := checkIdentExpr(s, expr)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.BuiltinInt)

	s = makeGlobalScope()
	expr = &IdentExpr{Tok: makeTok(10, 13), Name: "x"}
	typ = checkIdentExpr(s, expr)
	expectNthError(t, s, 0, "(10:13) variable 'x' was used before it was declared")
	expectBool(t, typ.IsError(), true)
}

func TestCheckNumberExpr(t *testing.T) {
	s := makeGlobalScope()
	expr := &NumberExpr{Tok: nop, Val: 123}
	typ := checkNumberExpr(s, expr)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.BuiltinInt)
}

func TestCheckStringExpr(t *testing.T) {
	s := makeGlobalScope()
	expr := &StringExpr{Tok: nop, Val: "abc"}
	typ := checkStringExpr(s, expr)
	expectNoErrors(t, s)
	expectEquivalentType(t, typ, types.BuiltinStr)
}

func TestCheckBooleanExpr(t *testing.T) {
	s := makeGlobalScope()
	expr := &BooleanExpr{Tok: nop, Val: true}
	typ := checkBooleanExpr(s, expr)
	expectNoErrors(t, s)
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

func expectConversion(t *testing.T, note TypeNote, exp string) {
	t.Helper()
	got := convertTypeNote(note)
	if got.String() != exp {
		t.Errorf("Expected '%s', got '%v'", exp, got)
	}
}

func makeTok(line int, col int) token {
	return token{Loc: Loc{Line: line, Col: col}}
}

func expectVariable(t *testing.T, s Scope, name string, exp types.Type) {
	t.Helper()
	if s.HasVariable(name) {
		got := s.GetVariableType(name)
		expectEquivalentType(t, got, exp)
	} else {
		t.Errorf("Expected variable '%s', none found", name)
	}
}

func expectLocalVariableType(t *testing.T, s Scope, name string, exp types.Type) {
	t.Helper()
	if s.HasLocalVariable(name) {
		got := s.GetLocalVariableType(name)
		expectEquivalentType(t, got, exp)
	} else {
		t.Errorf("Expected local variable '%s', none found", name)
	}
}
