package parser

import (
	"plaid/lexer"
	"testing"
)

var nop = lexer.Token{}

func TestProgram(t *testing.T) {
	(Program{}).isNode()

	prog := Program{[]Stmt{
		DeclarationStmt{nop, IdentExpr{nop, "a"}, NumberExpr{nop, 123}},
		DeclarationStmt{nop, IdentExpr{nop, "b"}, NumberExpr{nop, 456}},
	}}
	expectString(t, prog, "(let a 123)\n(let b 456)")

	prog = Program{[]Stmt{}}
	expectString(t, prog, "")
}

func TestStmtBlock(t *testing.T) {
	(StmtBlock{}).isNode()

	block := StmtBlock{nop, []Stmt{
		DeclarationStmt{nop, IdentExpr{nop, "a"}, NumberExpr{nop, 123}},
		DeclarationStmt{nop, IdentExpr{nop, "b"}, NumberExpr{nop, 456}},
	}, nop}
	expectString(t, block, "{\n  (let a 123)\n  (let b 456)}")
}

func TestIfStmt(t *testing.T) {
	(IfStmt{}).isNode()
	(IfStmt{}).isStmt()

	block := StmtBlock{nop, []Stmt{}, nop}
	expectString(t, IfStmt{nop, BooleanExpr{nop, true}, block}, "(if true {})")
}

func TestDeclarationStmt(t *testing.T) {
	(DeclarationStmt{}).isNode()
	(DeclarationStmt{}).isStmt()

	expectString(t, DeclarationStmt{nop, IdentExpr{nop, "a"}, NumberExpr{nop, 123}}, "(let a 123)")
}

func TestReturnStmt(t *testing.T) {
	(ReturnStmt{}).isNode()
	(ReturnStmt{}).isStmt()

	expectString(t, ReturnStmt{nop, nil}, "(return)")
	expectString(t, ReturnStmt{nop, NumberExpr{nop, 123}}, "(return 123)")
}

func TestExprStmt(t *testing.T) {
	(ExprStmt{}).isNode()
	(ExprStmt{}).isStmt()

	expectString(t, ExprStmt{IdentExpr{nop, "abc"}}, "abc")
}

func TestTypeNoteVoid(t *testing.T) {
	expectStart(t, TypeNoteVoid{}, 0, 0)
	expectString(t, TypeNoteVoid{}, "Void")
	(TypeNoteVoid{}).isNode()
	(TypeNoteVoid{}).isType()
}

func TestTypeNoteTuple(t *testing.T) {
	(TypeNoteTuple{}).isNode()
	(TypeNoteTuple{}).isType()

	expectString(t, TypeNoteTuple{}, "()")
	tuple := TypeNoteTuple{nop, []TypeNote{TypeNoteIdent{nop, "Bool"}, TypeNoteOptional{nop, TypeNoteIdent{nop, "Str"}}}}
	expectString(t, tuple, "(Bool Str?)")
}

func TestTypeNoteFunction(t *testing.T) {
	(TypeNoteFunction{}).isNode()
	(TypeNoteFunction{}).isType()

	args := TypeNoteTuple{}
	expectString(t, TypeNoteFunction{args, TypeNoteIdent{nop, "Int"}}, "() => Int")

	args = TypeNoteTuple{nop, []TypeNote{TypeNoteIdent{nop, "Bool"}, TypeNoteOptional{nop, TypeNoteIdent{nop, "Str"}}}}
	expectString(t, TypeNoteFunction{args, TypeNoteIdent{nop, "Int"}}, "(Bool Str?) => Int")
}

func TestTypeNoteIdent(t *testing.T) {
	(TypeNoteIdent{}).isNode()
	(TypeNoteIdent{}).isType()

	expectString(t, TypeNoteIdent{nop, "Int"}, "Int")
}

func TestTypeNoteList(t *testing.T) {
	(TypeNoteList{}).isNode()
	(TypeNoteList{}).isType()

	expectString(t, TypeNoteList{nop, TypeNoteIdent{nop, "Int"}}, "[Int]")
}

func TestTypeNoteOptional(t *testing.T) {
	(TypeNoteOptional{}).isNode()
	(TypeNoteOptional{}).isType()

	expectString(t, TypeNoteOptional{nop, TypeNoteIdent{nop, "Int"}}, "Int?")
}

func TestFunctionExpr(t *testing.T) {
	(FunctionExpr{}).isNode()
	(FunctionExpr{}).isExpr()
	(FunctionParam{}).isNode()

	params := []FunctionParam{
		FunctionParam{IdentExpr{nop, "x"}, TypeNoteIdent{nop, "Int"}},
		FunctionParam{IdentExpr{nop, "y"}, nil},
	}
	ret := TypeNoteIdent{nop, "Str"}
	block := StmtBlock{nop, []Stmt{
		DeclarationStmt{nop, IdentExpr{nop, "z"}, NumberExpr{nop, 123}},
	}, nop}

	expectString(t, FunctionExpr{nop, params, ret, block, nil}, "(fn (x:Int y):Str {\n  (let z 123)})")
}

func TestDispatchExpr(t *testing.T) {
	(DispatchExpr{}).isNode()
	(DispatchExpr{}).isExpr()

	callee := IdentExpr{nop, "callee"}
	args := []Expr{
		NumberExpr{nop, 123},
		NumberExpr{nop, 456},
	}

	expectString(t, DispatchExpr{callee, args}, "(callee (123 456))")
	expectString(t, DispatchExpr{callee, nil}, "(callee ())")
}

func TestAssignExpr(t *testing.T) {
	(AssignExpr{}).isNode()
	(AssignExpr{}).isExpr()

	expectString(t, AssignExpr{nop, IdentExpr{nop, "a"}, IdentExpr{nop, "b"}}, "(= a b)")
}

func TestListExpr(t *testing.T) {
	(ListExpr{}).isNode()
	(ListExpr{}).isExpr()

	expectString(t, ListExpr{nop, []Expr{}}, "[ ]")
	expectString(t, ListExpr{nop, []Expr{IdentExpr{nop, "a"}}}, "[ a ]")
}

func TestSubscriptExpr(t *testing.T) {
	(SubscriptExpr{}).isNode()
	(SubscriptExpr{}).isExpr()

	expectString(t, SubscriptExpr{IdentExpr{nop, "a"}, NumberExpr{nop, 0}}, "a[0]")
}

func TestBinaryExpr(t *testing.T) {
	(BinaryExpr{}).isNode()
	(BinaryExpr{}).isExpr()

	expectString(t, BinaryExpr{"+", nop, NumberExpr{nop, 123}, NumberExpr{nop, 456}}, "(+ 123 456)")
}

func TestUnaryExpr(t *testing.T) {
	(UnaryExpr{}).isNode()
	(UnaryExpr{}).isExpr()

	expectString(t, UnaryExpr{"+", nop, NumberExpr{nop, 123}}, "(+ 123)")
}

func TestIdentExpr(t *testing.T) {
	(IdentExpr{}).isNode()
	(IdentExpr{}).isExpr()

	expectString(t, IdentExpr{nop, "abc"}, "abc")
}

func TestStringExpr(t *testing.T) {
	(StringExpr{}).isNode()
	(StringExpr{}).isExpr()

	expectString(t, StringExpr{nop, "abc"}, "\"abc\"")
}

func TestNumberExpr(t *testing.T) {
	(NumberExpr{}).isNode()
	(NumberExpr{}).isExpr()

	expectString(t, NumberExpr{nop, 123}, "123")
}

func TestBooleanExpr(t *testing.T) {
	(BooleanExpr{}).isNode()
	(BooleanExpr{}).isExpr()

	expectString(t, BooleanExpr{nop, true}, "true")
	expectString(t, BooleanExpr{nop, false}, "false")
}

func TestIndentBlock(t *testing.T) {
	source := "foo\nbar\n  baz"
	exp := "...foo\n...bar\n...  baz"
	got := indentBlock("...", source)

	if exp != got {
		t.Errorf("Expected '%s', got '%s'\n", exp, got)
	}
}

func expectString(t *testing.T, node Node, exp string) {
	got := node.String()

	if exp != got {
		t.Errorf("Expected %s, got %s\n", exp, got)
	}
}
