package lexer

import (
	"testing"
)

func TestLexerPeekNext(t *testing.T) {
	var tok Token
	var err error
	lexer := Lex(Scan("abc def"))

	tok, err = lexer.Peek()
	if err != nil {
		t.Error(err)
	} else {
		expectToken(t, Token{Ident, "abc", Loc{1, 1}}, tok)
		expectToken(t, Token{Ident, "abc", Loc{1, 1}}, lexer.buffer[0])
	}

	tok, err = lexer.Peek()
	if err != nil {
		t.Error(err)
	} else {
		expectToken(t, Token{Ident, "abc", Loc{1, 1}}, tok)
		expectToken(t, Token{Ident, "abc", Loc{1, 1}}, lexer.buffer[0])
	}

	tok, err = lexer.Next()
	if err != nil {
		t.Error(err)
	} else {
		expectToken(t, Token{Ident, "abc", Loc{1, 1}}, tok)
	}

	tok, err = lexer.Peek()
	if err != nil {
		t.Error(err)
	} else {
		expectToken(t, Token{Ident, "def", Loc{1, 5}}, tok)
		expectToken(t, Token{Ident, "def", Loc{1, 5}}, lexer.buffer[0])
	}

	tok, err = lexer.Next()
	if err != nil {
		t.Error(err)
	} else {
		expectToken(t, Token{Ident, "def", Loc{1, 5}}, tok)
	}

	tok, err = lexer.Next()
	if err != nil {
		t.Error(err)
	} else {
		expectToken(t, Token{EOF, "", Loc{1, 7}}, tok)
	}

	tok, err = lexer.Next()
	if err != nil {
		t.Error(err)
	} else {
		expectToken(t, Token{EOF, "", Loc{1, 7}}, tok)
	}

	tok, err = lexer.Peek()
	if err != nil {
		t.Error(err)
	} else {
		expectToken(t, Token{EOF, "", Loc{1, 7}}, tok)
	}
}

func TestIsWhitespace(t *testing.T) {
	expectBool(t, isWhitespace, ' ', true)
	expectBool(t, isWhitespace, '\n', true)
	expectBool(t, isWhitespace, '\t', true)
	expectBool(t, isWhitespace, '\000', false)
	expectBool(t, isWhitespace, 'a', false)
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

func TestEatToken(t *testing.T) {
	expectLexer(t, eatToken, "foo", Token{Ident, "foo", Loc{1, 1}})

	expectLexerError(t, eatToken, "@", "(1:1) unexpected symbol")
}

func TestEatWordToken(t *testing.T) {
	expectLexer(t, eatWordToken, "foo", Token{Ident, "foo", Loc{1, 1}})

	expectLexerError(t, eatWordToken, "123", "(1:1) expected word")
	expectLexerError(t, eatWordToken, "", "(1:0) expected word")
}

type charPred func(rune) bool

func expectBool(t *testing.T, fn charPred, r rune, exp bool) {
	got := fn(r)

	if exp != got {
		t.Errorf("Expected %t, got %t\n", exp, got)
	}
}

type lexFunc func(scanner *CharBuffer) (Token, error)

func expectLexer(t *testing.T, fn lexFunc, source string, exp Token) {
	scanner := Scan(source)
	got, err := fn(scanner)

	if err != nil {
		t.Error(err)
	} else {
		expectToken(t, exp, got)
	}
}

func expectLexerError(t *testing.T, fn lexFunc, source string, msg string) {
	scanner := Scan(source)
	got, err := fn(scanner)

	if err == nil {
		t.Errorf("Expected syntax error '%s', got Token.typ %s\n", msg, got.typ)
	} else if msg != err.Error() {
		t.Errorf("Expected '%s', got '%s'", msg, err.Error())
	}
}

func expectToken(t *testing.T, exp Token, got Token) {
	if exp.typ != got.typ {
		t.Errorf("Expected Token.typ %s, got %s\n", exp.typ, got.typ)
	}

	if exp.lexeme != got.lexeme {
		t.Errorf("Expected Token.lexeme '%s', got '%s'\n", exp.lexeme, got.lexeme)
	}

	expectLoc(t, exp.loc, got.loc)
}
