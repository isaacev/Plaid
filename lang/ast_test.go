package lang

import (
	"testing"
)

var nop = token{}

func TestProgram(t *testing.T) {
	(RootNode{}).isNode()

	prog := RootNode{[]Stmt{
		DeclarationStmt{nop, &IdentExpr{nop, "a"}, &NumberExpr{nop, 123}},
		DeclarationStmt{nop, &IdentExpr{nop, "b"}, &NumberExpr{nop, 456}},
	}}
	expectASTString(t, prog, "(let a 123)\n(let b 456)")

	prog = RootNode{[]Stmt{}}
	expectASTString(t, prog, "")
}

func TestStmtBlock(t *testing.T) {
	(StmtBlock{}).isNode()

	block := StmtBlock{nop, []Stmt{
		DeclarationStmt{nop, &IdentExpr{nop, "a"}, &NumberExpr{nop, 123}},
		DeclarationStmt{nop, &IdentExpr{nop, "b"}, &NumberExpr{nop, 456}},
	}, nop}
	expectASTString(t, block, "{\n  (let a 123)\n  (let b 456)}")
}

func TestUseStmt(t *testing.T) {
	(UseStmt{}).isNode()
	(UseStmt{}).isStmt()

	path := &StringExpr{Val: "lib"}
	filter := []*UseFilter{}
	expectASTString(t, UseStmt{Path: path}, `(use "lib")`)
	expectASTString(t, UseStmt{Path: path, Filter: filter}, `(use "lib")`)
	expectStart(t, UseStmt{Path: path}, 0, 0)

	filter = append(filter, &UseFilter{&IdentExpr{Name: "fn1"}})
	expectASTString(t, UseStmt{Path: path, Filter: filter}, `(use "lib" (fn1))`)

	filter = append(filter, &UseFilter{&IdentExpr{Name: "fn2"}})
	expectASTString(t, UseStmt{Path: path, Filter: filter}, `(use "lib" (fn1 fn2))`)
}

func TestUseFilter(t *testing.T) {
	(UseFilter{}).isNode()

	filter := UseFilter{&IdentExpr{Name: "func1"}}
	expectASTString(t, filter, `func1`)
	expectStart(t, filter, 0, 0)
}

func TestPubStmt(t *testing.T) {
	(PubStmt{}).isNode()
	(PubStmt{}).isStmt()

	decl := &DeclarationStmt{Name: &IdentExpr{Name: "foo"}, Expr: &IdentExpr{Name: "bar"}}
	expectASTString(t, PubStmt{Stmt: decl}, `(pub (let foo bar))`)
	expectStart(t, PubStmt{Stmt: decl}, 0, 0)
}

func TestIfStmt(t *testing.T) {
	(IfStmt{}).isNode()
	(IfStmt{}).isStmt()

	block := &StmtBlock{nop, []Stmt{}, nop}
	expectASTString(t, IfStmt{nop, &BooleanExpr{nop, true}, block}, "(if true {})")
}

func TestDeclarationStmt(t *testing.T) {
	(DeclarationStmt{}).isNode()
	(DeclarationStmt{}).isStmt()

	expectASTString(t, DeclarationStmt{nop, &IdentExpr{nop, "a"}, &NumberExpr{nop, 123}}, "(let a 123)")
}

func TestReturnStmt(t *testing.T) {
	(ReturnStmt{}).isNode()
	(ReturnStmt{}).isStmt()

	expectASTString(t, ReturnStmt{nop, nil}, "(return)")
	expectASTString(t, ReturnStmt{nop, &NumberExpr{nop, 123}}, "(return 123)")
}

func TestExprStmt(t *testing.T) {
	(ExprStmt{}).isNode()
	(ExprStmt{}).isStmt()

	expectASTString(t, ExprStmt{IdentExpr{nop, "abc"}}, "abc")
}

func TestTypeNoteAny(t *testing.T) {
	expectStart(t, TypeNoteAny{}, 0, 0)
	expectASTString(t, TypeNoteAny{}, "Any")
	(TypeNoteAny{}).isNode()
	(TypeNoteAny{}).isType()
}

func TestTypeNoteVoid(t *testing.T) {
	expectStart(t, TypeNoteVoid{}, 0, 0)
	expectASTString(t, TypeNoteVoid{}, "Void")
	(TypeNoteVoid{}).isNode()
	(TypeNoteVoid{}).isType()
}

func TestTypeNoteTuple(t *testing.T) {
	(TypeNoteTuple{}).isNode()
	(TypeNoteTuple{}).isType()

	expectASTString(t, TypeNoteTuple{}, "()")
	tuple := TypeNoteTuple{nop, []TypeNote{TypeNoteIdent{nop, "Bool"}, TypeNoteOptional{nop, TypeNoteIdent{nop, "Str"}}}}
	expectASTString(t, tuple, "(Bool Str?)")
}

func TestTypeNoteFunction(t *testing.T) {
	(TypeNoteFunction{}).isNode()
	(TypeNoteFunction{}).isType()

	args := TypeNoteTuple{}
	expectASTString(t, TypeNoteFunction{args, TypeNoteIdent{nop, "Int"}}, "() => Int")

	args = TypeNoteTuple{nop, []TypeNote{TypeNoteIdent{nop, "Bool"}, TypeNoteOptional{nop, TypeNoteIdent{nop, "Str"}}}}
	expectASTString(t, TypeNoteFunction{args, TypeNoteIdent{nop, "Int"}}, "(Bool Str?) => Int")
}

func TestTypeNoteIdent(t *testing.T) {
	(TypeNoteIdent{}).isNode()
	(TypeNoteIdent{}).isType()

	expectASTString(t, TypeNoteIdent{nop, "Int"}, "Int")
}

func TestTypeNoteList(t *testing.T) {
	(TypeNoteList{}).isNode()
	(TypeNoteList{}).isType()

	expectASTString(t, TypeNoteList{nop, TypeNoteIdent{nop, "Int"}}, "[Int]")
}

func TestTypeNoteOptional(t *testing.T) {
	(TypeNoteOptional{}).isNode()
	(TypeNoteOptional{}).isType()

	expectASTString(t, TypeNoteOptional{nop, TypeNoteIdent{nop, "Int"}}, "Int?")
}

func TestFunctionExpr(t *testing.T) {
	(FunctionExpr{}).isNode()
	(FunctionExpr{}).isExpr()
	(FunctionParam{}).isNode()

	params := []*FunctionParam{
		&FunctionParam{&IdentExpr{nop, "x"}, TypeNoteIdent{nop, "Int"}},
		&FunctionParam{&IdentExpr{nop, "y"}, nil},
	}
	ret := TypeNoteIdent{nop, "Str"}
	block := &StmtBlock{nop, []Stmt{
		&DeclarationStmt{nop, &IdentExpr{nop, "z"}, &NumberExpr{nop, 123}},
	}, nop}

	expectASTString(t, &FunctionExpr{nop, params, ret, block}, "(fn (x:Int y):Str {\n  (let z 123)})")
}

func TestDispatchExpr(t *testing.T) {
	(DispatchExpr{}).isNode()
	(DispatchExpr{}).isExpr()

	callee := IdentExpr{nop, "callee"}
	args := []Expr{
		&NumberExpr{nop, 123},
		&NumberExpr{nop, 456},
	}

	expectASTString(t, DispatchExpr{callee, args}, "(callee (123 456))")
	expectASTString(t, DispatchExpr{callee, nil}, "(callee ())")
}

func TestAssignExpr(t *testing.T) {
	(AssignExpr{}).isNode()
	(AssignExpr{}).isExpr()

	expectASTString(t, AssignExpr{nop, &IdentExpr{nop, "a"}, &IdentExpr{nop, "b"}}, "(= a b)")
}

func TestListExpr(t *testing.T) {
	(ListExpr{}).isNode()
	(ListExpr{}).isExpr()

	expectASTString(t, ListExpr{nop, []Expr{}}, "[ ]")
	expectASTString(t, ListExpr{nop, []Expr{&IdentExpr{nop, "a"}}}, "[ a ]")
}

func TestSubscriptExpr(t *testing.T) {
	(SubscriptExpr{}).isNode()
	(SubscriptExpr{}).isExpr()

	expectASTString(t, &SubscriptExpr{&IdentExpr{nop, "a"}, &NumberExpr{nop, 0}}, "a[0]")
}

func TestAccessExpr(t *testing.T) {
	(AccessExpr{}).isNode()
	(AccessExpr{}).isExpr()

	expectASTString(t, &AccessExpr{&IdentExpr{nop, "a"}, &IdentExpr{nop, "b"}}, "(a).b")
}

func TestBinaryExpr(t *testing.T) {
	(BinaryExpr{}).isNode()
	(BinaryExpr{}).isExpr()

	expectASTString(t, BinaryExpr{"+", nop, &NumberExpr{nop, 123}, &NumberExpr{nop, 456}}, "(+ 123 456)")
}

func TestUnaryExpr(t *testing.T) {
	(UnaryExpr{}).isNode()
	(UnaryExpr{}).isExpr()

	expectASTString(t, UnaryExpr{"+", nop, &NumberExpr{nop, 123}}, "(+ 123)")
}

func TestIdentExpr(t *testing.T) {
	(IdentExpr{}).isNode()
	(IdentExpr{}).isExpr()

	expectASTString(t, &IdentExpr{nop, "abc"}, "abc")
}

func TestStringExpr(t *testing.T) {
	(StringExpr{}).isNode()
	(StringExpr{}).isExpr()

	expectASTString(t, StringExpr{nop, "abc"}, "\"abc\"")
}

func TestNumberExpr(t *testing.T) {
	(NumberExpr{}).isNode()
	(NumberExpr{}).isExpr()

	expectASTString(t, &NumberExpr{nop, 123}, "123")
}

func TestBooleanExpr(t *testing.T) {
	(BooleanExpr{}).isNode()
	(BooleanExpr{}).isExpr()

	expectASTString(t, BooleanExpr{nop, true}, "true")
	expectASTString(t, BooleanExpr{nop, false}, "false")
}

func TestIndentBlock(t *testing.T) {
	source := "foo\nbar\n  baz"
	exp := "...foo\n...bar\n...  baz"
	got := indentBlock("...", source)

	if exp != got {
		t.Errorf("Expected '%s', got '%s'\n", exp, got)
	}
}

func expectASTString(t *testing.T, node ASTNode, exp string) {
	t.Helper()
	got := node.String()

	if exp != got {
		t.Errorf("Expected %s, got %s\n", exp, got)
	}
}
