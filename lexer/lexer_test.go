package lexer

import (
	"fmt"
	"testing"
)

func TestLexerPeekNext(t *testing.T) {
	lexer := Lex(Scan("abc def"))

	expectToken(t, Token{Ident, "abc", Loc{1, 1}}, lexer.Peek())
	expectToken(t, Token{Ident, "abc", Loc{1, 1}}, lexer.buffer[0])
	expectToken(t, Token{Ident, "abc", Loc{1, 1}}, lexer.Peek())
	expectToken(t, Token{Ident, "abc", Loc{1, 1}}, lexer.buffer[0])
	expectToken(t, Token{Ident, "abc", Loc{1, 1}}, lexer.Next())
	expectToken(t, Token{Ident, "def", Loc{1, 5}}, lexer.Peek())
	expectToken(t, Token{Ident, "def", Loc{1, 5}}, lexer.buffer[0])
	expectToken(t, Token{Ident, "def", Loc{1, 5}}, lexer.Next())
	expectToken(t, Token{EOF, "", Loc{1, 7}}, lexer.Next())
	expectToken(t, Token{EOF, "", Loc{1, 7}}, lexer.Next())
	expectToken(t, Token{EOF, "", Loc{1, 7}}, lexer.Peek())
}

func TestIsWhitespace(t *testing.T) {
	expectBool(t, isWhitespace, ' ', true)
	expectBool(t, isWhitespace, '\n', true)
	expectBool(t, isWhitespace, '\t', true)
	expectBool(t, isWhitespace, '\000', false)
	expectBool(t, isWhitespace, 'a', false)
}

func TestIsOperator(t *testing.T) {
	expectBool(t, isOperator, '+', true)
	expectBool(t, isOperator, '-', true)
	expectBool(t, isOperator, '*', true)
	expectBool(t, isOperator, '/', true)
	expectBool(t, isOperator, '#', false)
}

func TestIsLetter(t *testing.T) {
	expectBool(t, isLetter, '`', false)
	expectBool(t, isLetter, 'a', true)
	expectBool(t, isLetter, 'z', true)
	expectBool(t, isLetter, '{', false)
	expectBool(t, isLetter, '@', false)
	expectBool(t, isLetter, 'A', true)
	expectBool(t, isLetter, 'Z', true)
	expectBool(t, isLetter, '[', false)
}

func TestIsDigit(t *testing.T) {
	expectBool(t, isDigit, '0', true)
	expectBool(t, isDigit, '9', true)
	expectBool(t, isDigit, '/', false)
	expectBool(t, isDigit, ':', false)
}

func TestEatToken(t *testing.T) {
	expectLexer(t, eatToken, "", Token{EOF, "", Loc{1, 0}})
	expectLexer(t, eatToken, "  \nfoo", Token{Ident, "foo", Loc{2, 1}})
	expectLexer(t, eatToken, "+", Token{Plus, "+", Loc{1, 1}})
	expectLexer(t, eatToken, "-", Token{Dash, "-", Loc{1, 1}})
	expectLexer(t, eatToken, "*", Token{Star, "*", Loc{1, 1}})
	expectLexer(t, eatToken, "/", Token{Slash, "/", Loc{1, 1}})
	expectLexer(t, eatToken, "foo", Token{Ident, "foo", Loc{1, 1}})
	expectLexer(t, eatToken, "123", Token{Number, "123", Loc{1, 1}})

	expectLexerError(t, eatToken, "@", "(1:1) unexpected symbol")
}

func TestEatOperatorToken(t *testing.T) {
	expectLexer(t, eatOperatorToken, "+", Token{Plus, "+", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "-", Token{Dash, "-", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "*", Token{Star, "*", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "/", Token{Slash, "/", Loc{1, 1}})

	expectLexerError(t, eatOperatorToken, "@", "(1:1) expected operator")
}

func TestEatWordToken(t *testing.T) {
	expectLexer(t, eatWordToken, "foo", Token{Ident, "foo", Loc{1, 1}})

	expectLexerError(t, eatWordToken, "123", "(1:1) expected word")
	expectLexerError(t, eatWordToken, "", "(1:0) expected word")
}

func TestEatNumberToken(t *testing.T) {
	expectLexer(t, eatNumberToken, "123", Token{Number, "123", Loc{1, 1}})

	expectLexerError(t, eatNumberToken, "foo", "(1:1) expected number")
	expectLexerError(t, eatNumberToken, "", "(1:0) expected number")
}

type charPred func(rune) bool

func expectBool(t *testing.T, fn charPred, r rune, exp bool) {
	got := fn(r)

	if exp != got {
		t.Errorf("Expected %t, got %t\n", exp, got)
	}
}

type lexFunc func(scanner *Scanner) Token

func expectLexer(t *testing.T, fn lexFunc, source string, exp Token) {
	scanner := Scan(source)
	got := fn(scanner)
	expectToken(t, exp, got)
}

func expectLexerError(t *testing.T, fn lexFunc, source string, msg string) {
	scanner := Scan(source)
	got := fn(scanner)

	if got.Type == Error {
		if msg != formatErrorMessage(got) {
			t.Errorf("Expected syntax error '%s', got '%s'\n", msg, formatErrorMessage(got))
		}
	} else {
		t.Errorf("Expected Error, got %v\n", got)
	}
}

func expectToken(t *testing.T, exp Token, got Token) {
	if exp.Type != got.Type {
		t.Errorf("Expected Token.Type %s, got %s\n", exp.Type, got.Type)
	}

	if exp.Lexeme != got.Lexeme {
		t.Errorf("Expected Token.Lexeme '%s', got '%s'\n", exp.Lexeme, got.Lexeme)
	}

	expectLoc(t, exp.Loc, got.Loc)
}

func formatErrorMessage(tok Token) string {
	return fmt.Sprintf("%s %s", tok.Loc, tok.Lexeme)
}
