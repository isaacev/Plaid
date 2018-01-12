package lang

import (
	"fmt"
	"testing"
)

func TestLexerPeekNext(t *testing.T) {
	lexer := Lex("", Scan("abc def"))

	expectToken(t, Token{TokIdent, "abc", Loc{1, 1}}, lexer.Peek())
	expectToken(t, Token{TokIdent, "abc", Loc{1, 1}}, lexer.buffer[0])
	expectToken(t, Token{TokIdent, "abc", Loc{1, 1}}, lexer.Peek())
	expectToken(t, Token{TokIdent, "abc", Loc{1, 1}}, lexer.buffer[0])
	expectToken(t, Token{TokIdent, "abc", Loc{1, 1}}, lexer.Next())
	expectToken(t, Token{TokIdent, "def", Loc{1, 5}}, lexer.Peek())
	expectToken(t, Token{TokIdent, "def", Loc{1, 5}}, lexer.buffer[0])
	expectToken(t, Token{TokIdent, "def", Loc{1, 5}}, lexer.Next())
	expectToken(t, Token{TokEOF, "", Loc{1, 7}}, lexer.Next())
	expectToken(t, Token{TokEOF, "", Loc{1, 7}}, lexer.Next())
	expectToken(t, Token{TokEOF, "", Loc{1, 7}}, lexer.Peek())
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
	expectLexer(t, eatToken, "", Token{TokEOF, "", Loc{1, 0}})
	expectLexer(t, eatToken, "  \nfoo", Token{TokIdent, "foo", Loc{2, 1}})
	expectLexer(t, eatToken, "+", Token{TokPlus, "+", Loc{1, 1}})
	expectLexer(t, eatToken, "-", Token{TokDash, "-", Loc{1, 1}})
	expectLexer(t, eatToken, "*", Token{TokStar, "*", Loc{1, 1}})
	expectLexer(t, eatToken, "/", Token{TokSlash, "/", Loc{1, 1}})
	expectLexer(t, eatToken, "?", Token{TokQuestion, "?", Loc{1, 1}})
	expectLexer(t, eatToken, ";", Token{TokSemi, ";", Loc{1, 1}})
	expectLexer(t, eatToken, ",", Token{TokComma, ",", Loc{1, 1}})
	expectLexer(t, eatToken, "(", Token{TokParenL, "(", Loc{1, 1}})
	expectLexer(t, eatToken, ")", Token{TokParenR, ")", Loc{1, 1}})
	expectLexer(t, eatToken, "{", Token{TokBraceL, "{", Loc{1, 1}})
	expectLexer(t, eatToken, "}", Token{TokBraceR, "}", Loc{1, 1}})
	expectLexer(t, eatToken, "[", Token{TokBracketL, "[", Loc{1, 1}})
	expectLexer(t, eatToken, "]", Token{TokBracketR, "]", Loc{1, 1}})
	expectLexer(t, eatToken, "foo", Token{TokIdent, "foo", Loc{1, 1}})
	expectLexer(t, eatToken, "fn", Token{TokFn, "fn", Loc{1, 1}})
	expectLexer(t, eatToken, "if", Token{TokIf, "if", Loc{1, 1}})
	expectLexer(t, eatToken, "let", Token{TokLet, "let", Loc{1, 1}})
	expectLexer(t, eatToken, "return", Token{TokReturn, "return", Loc{1, 1}})
	expectLexer(t, eatToken, "self", Token{TokSelf, "self", Loc{1, 1}})
	expectLexer(t, eatToken, "use", Token{TokUse, "use", Loc{1, 1}})
	expectLexer(t, eatToken, "pub", Token{TokPub, "pub", Loc{1, 1}})
	expectLexer(t, eatToken, "true", Token{TokBoolean, "true", Loc{1, 1}})
	expectLexer(t, eatToken, "false", Token{TokBoolean, "false", Loc{1, 1}})
	expectLexer(t, eatToken, "123", Token{TokNumber, "123", Loc{1, 1}})
	expectLexer(t, eatToken, `"foo"`, Token{TokString, `"foo"`, Loc{1, 1}})

	expectLexerError(t, eatToken, "@", "(1:1) unexpected symbol")
}

func TestEatOperatorToken(t *testing.T) {
	expectLexer(t, eatOperatorToken, "+", Token{TokPlus, "+", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "-", Token{TokDash, "-", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "*", Token{TokStar, "*", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "/", Token{TokSlash, "/", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "?", Token{TokQuestion, "?", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, ":", Token{TokColon, ":", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, ":=", Token{TokAssign, ":=", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "=>", Token{TokArrow, "=>", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "<", Token{TokLT, "<", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, "<=", Token{TokLTEquals, "<=", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, ">", Token{TokGT, ">", Loc{1, 1}})
	expectLexer(t, eatOperatorToken, ">=", Token{TokGTEquals, ">=", Loc{1, 1}})

	expectLexerError(t, eatOperatorToken, "@", "(1:1) expected operator")
	expectLexerError(t, eatOperatorToken, "=", "(1:1) expected operator")
}

func TestEatSemicolonToken(t *testing.T) {
	expectLexer(t, eatSemicolonToken, ";", Token{TokSemi, ";", Loc{1, 1}})

	expectLexerError(t, eatSemicolonToken, "@", "(1:1) expected semicolon")
}

func TestEatCommaToken(t *testing.T) {
	expectLexer(t, eatCommaToken, ",", Token{TokComma, ",", Loc{1, 1}})

	expectLexerError(t, eatCommaToken, "@", "(1:1) expected comma")
}

func TestEatParen(t *testing.T) {
	expectLexer(t, eatParenToken, "(", Token{TokParenL, "(", Loc{1, 1}})
	expectLexer(t, eatParenToken, ")", Token{TokParenR, ")", Loc{1, 1}})

	expectLexerError(t, eatParenToken, "@", "(1:1) expected paren")
}

func TestEatBrace(t *testing.T) {
	expectLexer(t, eatBraceToken, "{", Token{TokBraceL, "{", Loc{1, 1}})
	expectLexer(t, eatBraceToken, "}", Token{TokBraceR, "}", Loc{1, 1}})

	expectLexerError(t, eatBraceToken, "@", "(1:1) expected brace")
}

func TestEatBracket(t *testing.T) {
	expectLexer(t, eatBracketToken, "[", Token{TokBracketL, "[", Loc{1, 1}})
	expectLexer(t, eatBracketToken, "]", Token{TokBracketR, "]", Loc{1, 1}})

	expectLexerError(t, eatBracketToken, "@", "(1:1) expected bracket")
}

func TestEatWordToken(t *testing.T) {
	expectLexer(t, eatWordToken, "foo", Token{TokIdent, "foo", Loc{1, 1}})
	expectLexer(t, eatWordToken, "fn", Token{TokFn, "fn", Loc{1, 1}})
	expectLexer(t, eatWordToken, "if", Token{TokIf, "if", Loc{1, 1}})
	expectLexer(t, eatWordToken, "let", Token{TokLet, "let", Loc{1, 1}})
	expectLexer(t, eatWordToken, "return", Token{TokReturn, "return", Loc{1, 1}})
	expectLexer(t, eatWordToken, "self", Token{TokSelf, "self", Loc{1, 1}})
	expectLexer(t, eatWordToken, "use", Token{TokUse, "use", Loc{1, 1}})
	expectLexer(t, eatWordToken, "pub", Token{TokPub, "pub", Loc{1, 1}})

	expectLexerError(t, eatWordToken, "123", "(1:1) expected word")
	expectLexerError(t, eatWordToken, "", "(1:0) expected word")
}

func TestEatNumberToken(t *testing.T) {
	expectLexer(t, eatNumberToken, "123", Token{TokNumber, "123", Loc{1, 1}})

	expectLexerError(t, eatNumberToken, "foo", "(1:1) expected number")
	expectLexerError(t, eatNumberToken, "", "(1:0) expected number")
}

func TestEatStringToken(t *testing.T) {
	expectLexer(t, eatStringToken, `"foo"`, Token{TokString, `"foo"`, Loc{1, 1}})

	expectLexerError(t, eatStringToken, "123", "(1:1) expected string")
	expectLexerError(t, eatStringToken, `"foo`, "(1:4) unclosed string")
	expectLexerError(t, eatStringToken, "\"foo\n\"", "(1:5) unclosed string")
}

type charPred func(rune) bool

func expectRunePredicateBool(t *testing.T, fn charPred, r rune, exp bool) {
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

	if got.Type == TokError {
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
