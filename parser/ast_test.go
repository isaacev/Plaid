package parser

import (
	"plaid/lexer"
	"testing"
)

var tok = lexer.Token{}

func TestDeclarationStmt(t *testing.T) {
	(DeclarationStmt{}).isNode()
	(DeclarationStmt{}).isStmt()

	expectString(t, DeclarationStmt{tok, IdentExpr{tok, "a"}, NumberExpr{tok, 123}}, "(let a 123)")
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
