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

func TestTypeTuple(t *testing.T) {
	(TypeTuple{}).isNode()
	(TypeTuple{}).isType()

	expectString(t, TypeTuple{}, "()")
	tuple := TypeTuple{nop, []TypeSig{TypeIdent{nop, "Bool"}, TypeOptional{nop, TypeIdent{nop, "Str"}}}}
	expectString(t, tuple, "(Bool Str?)")
}

func TestTypeFunction(t *testing.T) {
	(TypeFunction{}).isNode()
	(TypeFunction{}).isType()

	args := TypeTuple{}
	expectString(t, TypeFunction{args, TypeIdent{nop, "Int"}}, "() => Int")

	args = TypeTuple{nop, []TypeSig{TypeIdent{nop, "Bool"}, TypeOptional{nop, TypeIdent{nop, "Str"}}}}
	expectString(t, TypeFunction{args, TypeIdent{nop, "Int"}}, "(Bool Str?) => Int")
}

func TestTypeIdent(t *testing.T) {
	(TypeIdent{}).isNode()
	(TypeIdent{}).isType()

	expectString(t, TypeIdent{nop, "Int"}, "Int")
}

func TestTypeList(t *testing.T) {
	(TypeList{}).isNode()
	(TypeList{}).isType()

	expectString(t, TypeList{nop, TypeIdent{nop, "Int"}}, "[Int]")
}

func TestTypeOptional(t *testing.T) {
	(TypeOptional{}).isNode()
	(TypeOptional{}).isType()

	expectString(t, TypeOptional{nop, TypeIdent{nop, "Int"}}, "Int?")
}

func TestFunctionExpr(t *testing.T) {
	(FunctionExpr{}).isNode()
	(FunctionExpr{}).isExpr()
	(FunctionParam{}).isNode()

	params := []FunctionParam{
		FunctionParam{IdentExpr{nop, "x"}, TypeIdent{nop, "Int"}},
		FunctionParam{IdentExpr{nop, "y"}, nil},
	}
	ret := TypeIdent{nop, "Str"}
	block := StmtBlock{nop, []Stmt{
		DeclarationStmt{nop, IdentExpr{nop, "z"}, NumberExpr{nop, 123}},
	}, nop}

	expectString(t, FunctionExpr{nop, params, ret, block}, "(fn (x:Int y):Str {\n  (let z 123)})")
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
