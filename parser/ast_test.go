package parser

import (
	"plaid/lexer"
	"testing"
)

var tok = lexer.Token{}

func TestProgram(t *testing.T) {
	(Program{}).isNode()

	prog := Program{[]Stmt{
		DeclarationStmt{tok, IdentExpr{tok, "a"}, NumberExpr{tok, 123}},
		DeclarationStmt{tok, IdentExpr{tok, "b"}, NumberExpr{tok, 456}},
	}}
	expectString(t, prog, "(let a 123)\n(let b 456)")

	prog = Program{[]Stmt{}}
	expectString(t, prog, "")
}

func TestStmtBlock(t *testing.T) {
	(StmtBlock{}).isNode()

	block := StmtBlock{tok, []Stmt{
		DeclarationStmt{tok, IdentExpr{tok, "a"}, NumberExpr{tok, 123}},
		DeclarationStmt{tok, IdentExpr{tok, "b"}, NumberExpr{tok, 456}},
	}, tok}
	expectString(t, block, "{\n  (let a 123)\n  (let b 456)}")
}

func TestDeclarationStmt(t *testing.T) {
	(DeclarationStmt{}).isNode()
	(DeclarationStmt{}).isStmt()

	expectString(t, DeclarationStmt{tok, IdentExpr{tok, "a"}, NumberExpr{tok, 123}}, "(let a 123)")
}

func TestTypeIdent(t *testing.T) {
	(TypeIdent{}).isNode()
	(TypeIdent{}).isType()

	expectString(t, TypeIdent{tok, "Int"}, "Int")
}

func TestTypeList(t *testing.T) {
	(TypeList{}).isNode()
	(TypeList{}).isType()

	expectString(t, TypeList{tok, TypeIdent{tok, "Int"}}, "[Int]")
}

func TestTypeOptional(t *testing.T) {
	(TypeOptional{}).isNode()
	(TypeOptional{}).isType()

	expectString(t, TypeOptional{tok, TypeIdent{tok, "Int"}}, "Int?")
}

func TestFunctionExpr(t *testing.T) {
	(FunctionExpr{}).isNode()
	(FunctionExpr{}).isExpr()

	params := []FunctionParam{
		FunctionParam{IdentExpr{tok, "x"}, TypeIdent{tok, "Int"}},
		FunctionParam{IdentExpr{tok, "y"}, TypeIdent{tok, "Bool"}},
	}
	ret := TypeIdent{tok, "Str"}
	block := StmtBlock{tok, []Stmt{
		DeclarationStmt{tok, IdentExpr{tok, "z"}, NumberExpr{tok, 123}},
	}, tok}

	expectString(t, FunctionExpr{tok, params, ret, block}, "(fn (x:Int y:Bool):Str {\n  (let z 123)})")
}

func TestBinaryExpr(t *testing.T) {
	(BinaryExpr{}).isNode()
	(BinaryExpr{}).isExpr()

	expectString(t, BinaryExpr{"+", tok, NumberExpr{tok, 123}, NumberExpr{tok, 456}}, "(+ 123 456)")
}

func TestUnaryExpr(t *testing.T) {
	(UnaryExpr{}).isNode()
	(UnaryExpr{}).isExpr()

	expectString(t, UnaryExpr{"+", tok, NumberExpr{tok, 123}}, "(+ 123)")
}

func TestIdentExpr(t *testing.T) {
	(IdentExpr{}).isNode()
	(IdentExpr{}).isExpr()

	expectString(t, IdentExpr{tok, "abc"}, "abc")
}

func TestStringExpr(t *testing.T) {
	(StringExpr{}).isNode()
	(StringExpr{}).isExpr()

	expectString(t, StringExpr{tok, "abc"}, "\"abc\"")
}

func TestNumberExpr(t *testing.T) {
	(NumberExpr{}).isNode()
	(NumberExpr{}).isExpr()

	expectString(t, NumberExpr{tok, 123}, "123")
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
