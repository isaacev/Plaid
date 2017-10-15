package parser

import (
	"plaid/lexer"
	"testing"
)

func TestSyntaxError(t *testing.T) {
	tok := lexer.Token{Type: lexer.Error, Lexeme: "custom lexer error", Loc: lexer.Loc{Line: 2, Col: 5}}
	exp := "(2:5) custom lexer error"
	err := makeSyntaxError(tok, "generic error", true)
	if err.Error() != exp {
		t.Errorf("Expected '%s', got '%s'\n", exp, err.Error())
	}

	tok = lexer.Token{Type: lexer.Error, Lexeme: "generic error", Loc: lexer.Loc{Line: 2, Col: 5}}
	exp = "(2:5) custom parser error"
	err = makeSyntaxError(tok, "custom parser error", false)
	if err.Error() != exp {
		t.Errorf("Expected '%s', got '%s'\n", exp, err.Error())
	}
}
