package lang

import (
	"fmt"
	"testing"
)

func TestLexerPeekNext(t *testing.T) {
	lexer := makeLexer("", scan("abc def"))

	expectToken(t, token{tokIdent, "abc", Loc{1, 1}}, lexer.peek())
	expectToken(t, token{tokIdent, "abc", Loc{1, 1}}, lexer.buffer[0])
	expectToken(t, token{tokIdent, "abc", Loc{1, 1}}, lexer.peek())
	expectToken(t, token{tokIdent, "abc", Loc{1, 1}}, lexer.buffer[0])
	expectToken(t, token{tokIdent, "abc", Loc{1, 1}}, lexer.next())
	expectToken(t, token{tokIdent, "def", Loc{1, 5}}, lexer.peek())
	expectToken(t, token{tokIdent, "def", Loc{1, 5}}, lexer.buffer[0])
	expectToken(t, token{tokIdent, "def", Loc{1, 5}}, lexer.next())
	expectToken(t, token{tokEOF, "", Loc{1, 7}}, lexer.next())
	expectToken(t, token{tokEOF, "", Loc{1, 7}}, lexer.next())
	expectToken(t, token{tokEOF, "", Loc{1, 7}}, lexer.peek())
}

func TestIsWhitespace(t *testing.T) {
	expectRunePredicateBool(t, isWhitespace, ' ', true)
	expectRunePredicateBool(t, isWhitespace, '\n', true)
	expectRunePredicateBool(t, isWhitespace, '\t', true)
	expectRunePredicateBool(t, isWhitespace, '\000', false)
	expectRunePredicateBool(t, isWhitespace, 'a', false)
}

func TestIsOperator(t *testing.T) {
	expectRunePredicateBool(t, isOperator, '+', true)
	expectRunePredicateBool(t, isOperator, '-', true)
	expectRunePredicateBool(t, isOperator, '*', true)
	expectRunePredicateBool(t, isOperator, '/', true)
	expectRunePredicateBool(t, isOperator, '?', true)
	expectRunePredicateBool(t, isOperator, ':', true)
	expectRunePredicateBool(t, isOperator, '=', true)
	expectRunePredicateBool(t, isOperator, '<', true)
	expectRunePredicateBool(t, isOperator, '>', true)
	expectRunePredicateBool(t, isOperator, '#', false)
}

func TestIsParen(t *testing.T) {
	expectRunePredicateBool(t, isParen, '(', true)
	expectRunePredicateBool(t, isParen, ')', true)
	expectRunePredicateBool(t, isParen, '\'', false)
	expectRunePredicateBool(t, isParen, '*', false)
}

func TestIsLetter(t *testing.T) {
	expectRunePredicateBool(t, isLetter, '`', false)
	expectRunePredicateBool(t, isLetter, 'a', true)
	expectRunePredicateBool(t, isLetter, 'z', true)
	expectRunePredicateBool(t, isLetter, '{', false)
	expectRunePredicateBool(t, isLetter, '@', false)
	expectRunePredicateBool(t, isLetter, 'A', true)
	expectRunePredicateBool(t, isLetter, 'Z', true)
	expectRunePredicateBool(t, isLetter, '[', false)
}

func TestIsDigit(t *testing.T) {
	expectRunePredicateBool(t, isDigit, '0', true)
	expectRunePredicateBool(t, isDigit, '9', true)
	expectRunePredicateBool(t, isDigit, '/', false)
	expectRunePredicateBool(t, isDigit, ':', false)
}

func TestIsDoubleQuote(t *testing.T) {
	expectRunePredicateBool(t, isDoubleQuote, '"', true)
	expectRunePredicateBool(t, isDoubleQuote, '!', false)
	expectRunePredicateBool(t, isDoubleQuote, '#', false)
}

func TestEatToken(t *testing.T) {
	expectLexer(t, eatToken, "", token{tokEOF, "", Loc{1, 0}})
	expectLexer(t, eatToken, "  \nfoo", token{tokIdent, "foo", Loc{2, 1}})
	expectLexer(t, eatToken, "+", token{tokPlus, "+", Loc{1, 1}})
	expectLexer(t, eatToken, "-", token{tokDash, "-", Loc{1, 1}})
	expectLexer(t, eatToken, "*", token{tokStar, "*", Loc{1, 1}})
	expectLexer(t, eatToken, "/", token{tokSlash, "/", Loc{1, 1}})
	expectLexer(t, eatToken, "?", token{tokQuestion, "?", Loc{1, 1}})
	expectLexer(t, eatToken, ";", token{tokSemi, ";", Loc{1, 1}})
	expectLexer(t, eatToken, ",", token{tokComma, ",", Loc{1, 1}})
	expectLexer(t, eatToken, "(", token{tokParenL, "(", Loc{1, 1}})
	expectLexer(t, eatToken, ")", token{tokParenR, ")", Loc{1, 1}})
	expectLexer(t, eatToken, "{", token{tokBraceL, "{", Loc{1, 1}})
	expectLexer(t, eatToken, "}", token{tokBraceR, "}", Loc{1, 1}})
	expectLexer(t, eatToken, "[", token{tokBracketL, "[", Loc{1, 1}})
	expectLexer(t, eatToken, "]", token{tokBracketR, "]", Loc{1, 1}})
	expectLexer(t, eatToken, "foo", token{tokIdent, "foo", Loc{1, 1}})
	expectLexer(t, eatToken, "fn", token{tokFn, "fn", Loc{1, 1}})
	expectLexer(t, eatToken, "if", token{tokIf, "if", Loc{1, 1}})
	expectLexer(t, eatToken, "let", token{tokLet, "let", Loc{1, 1}})
	expectLexer(t, eatToken, "return", token{tokReturn, "return", Loc{1, 1}})
	expectLexer(t, eatToken, "self", token{tokSelf, "self", Loc{1, 1}})
	expectLexer(t, eatToken, "use", token{tokUse, "use", Loc{1, 1}})
	expectLexer(t, eatToken, "pub", token{tokPub, "pub", Loc{1, 1}})
	expectLexer(t, eatToken, "true", token{tokBoolean, "true", Loc{1, 1}})
	expectLexer(t, eatToken, "false", token{tokBoolean, "false", Loc{1, 1}})
	expectLexer(t, eatToken, "123", token{tokNumber, "123", Loc{1, 1}})
	expectLexer(t, eatToken, `"foo"`, token{tokString, `"foo"`, Loc{1, 1}})

	expectLexerError(t, eatToken, "@", "(1:1) unexpected symbol")
}

func TestEatOperatorToken(t *testing.T) {
	expectLexer(t, eatOperatorToken, "+", token{tokPlus, "+", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "-", token{tokDash, "-", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "*", token{tokStar, "*", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "/", token{tokSlash, "/", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "?", token{tokQuestion, "?", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, ":", token{tokColon, ":", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, ":=", token{tokAssign, ":=", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "=>", token{tokArrow, "=>", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "<", token{tokLT, "<", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "<=", token{tokLTEquals, "<=", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, ">", token{tokGT, ">", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, ">=", token{tokGTEquals, ">=", Loc{1, 1}})

	expectLexerError(t, eatOperatorToken, "@", "(1:1) expected operator")
	expectLexerError(t, eatOperatorToken, "=", "(1:1) expected operator")
}

func TestEatSemicolonToken(t *testing.T) {
	expectLexer(t, eatSemicolonToken, ";", token{tokSemi, ";", Loc{1, 1}})

	expectLexerError(t, eatSemicolonToken, "@", "(1:1) expected semicolon")
}

func TestEatCommaToken(t *testing.T) {
	expectLexer(t, eatCommaToken, ",", token{tokComma, ",", Loc{1, 1}})

	expectLexerError(t, eatCommaToken, "@", "(1:1) expected comma")
}

func TestEatParen(t *testing.T) {
	expectLexer(t, eatParenToken, "(", token{tokParenL, "(", Loc{1, 1}})
	expectLexer(t, eatParenToken, ")", token{tokParenR, ")", Loc{1, 1}})

	expectLexerError(t, eatParenToken, "@", "(1:1) expected paren")
}

func TestEatBrace(t *testing.T) {
	expectLexer(t, eatBraceToken, "{", token{tokBraceL, "{", Loc{1, 1}})
	expectLexer(t, eatBraceToken, "}", token{tokBraceR, "}", Loc{1, 1}})

	expectLexerError(t, eatBraceToken, "@", "(1:1) expected brace")
}

func TestEatBracket(t *testing.T) {
	expectLexer(t, eatBracketToken, "[", token{tokBracketL, "[", Loc{1, 1}})
	expectLexer(t, eatBracketToken, "]", token{tokBracketR, "]", Loc{1, 1}})

	expectLexerError(t, eatBracketToken, "@", "(1:1) expected bracket")
}

func TestEatWordToken(t *testing.T) {
	expectLexer(t, eatWordToken, "foo", token{tokIdent, "foo", Loc{1, 1}})
	expectLexer(t, eatWordToken, "fn", token{tokFn, "fn", Loc{1, 1}})
	expectLexer(t, eatWordToken, "if", token{tokIf, "if", Loc{1, 1}})
	expectLexer(t, eatWordToken, "let", token{tokLet, "let", Loc{1, 1}})
	expectLexer(t, eatWordToken, "return", token{tokReturn, "return", Loc{1, 1}})
	expectLexer(t, eatWordToken, "self", token{tokSelf, "self", Loc{1, 1}})
	expectLexer(t, eatWordToken, "use", token{tokUse, "use", Loc{1, 1}})
	expectLexer(t, eatWordToken, "pub", token{tokPub, "pub", Loc{1, 1}})

	expectLexerError(t, eatWordToken, "123", "(1:1) expected word")
	expectLexerError(t, eatWordToken, "", "(1:0) expected word")
}

func TestEatNumberToken(t *testing.T) {
	expectLexer(t, eatNumberToken, "123", token{tokNumber, "123", Loc{1, 1}})

	expectLexerError(t, eatNumberToken, "foo", "(1:1) expected number")
	expectLexerError(t, eatNumberToken, "", "(1:0) expected number")
}

func TestEatStringToken(t *testing.T) {
	expectLexer(t, eatStringToken, `"foo"`, token{tokString, `"foo"`, Loc{1, 1}})

	expectLexerError(t, eatStringToken, "123", "(1:1) expected string")
	expectLexerError(t, eatStringToken, `"foo`, "(1:4) unclosed string")
	expectLexerError(t, eatStringToken, "\"foo\n\"", "(1:5) unclosed string")
}

type charPred func(rune) bool

func expectRunePredicateBool(t *testing.T, fn charPred, r rune, exp bool) {
	t.Helper()
	got := fn(r)

	if exp != got {
		t.Errorf("Expected %t, got %t\n", exp, got)
	}
}

type lexFunc func(scn *scanner) token

func expectLexer(t *testing.T, fn lexFunc, source string, exp token) {
	t.Helper()
	scn := scan(source)
	got := fn(scn)
	expectToken(t, exp, got)
}

func expectLexerError(t *testing.T, fn lexFunc, source string, msg string) {
	t.Helper()
	scn := scan(source)
	got := fn(scn)

	if got.Type == tokError {
		if msg != formatErrorMessage(got) {
			t.Errorf("Expected syntax error '%s', got '%s'\n", msg, formatErrorMessage(got))
		}
	} else {
		t.Errorf("Expected Error, got %v\n", got)
	}
}

func expectToken(t *testing.T, exp token, got token) {
	t.Helper()
	if exp.Type != got.Type {
		t.Errorf("Expected Token.Type %s, got %s\n", exp.Type, got.Type)
	}

	if exp.Lexeme != got.Lexeme {
		t.Errorf("Expected Token.Lexeme '%s', got '%s'\n", exp.Lexeme, got.Lexeme)
	}

	expectLoc(t, exp.Loc, got.Loc)
}

func formatErrorMessage(tok token) string {
	return fmt.Sprintf("%s %s", tok.Loc, tok.Lexeme)
}
