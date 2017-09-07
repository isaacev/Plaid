package lexer

import (
	"fmt"
)

// Type classifies the different Tokens
type Type string

// Token classifications
const (
	EOF    Type = "EOF"
	Ident       = "Ident"
	Number      = "Number"
	String      = "String"
)

// Token is a basic syntactic unit
type Token struct {
	typ    Type
	lexeme string
	loc    Loc
}

// Lexer contains methods for generating a sequence of Tokens
type Lexer struct {
	buffer  []Token
	scanner *Scanner
}

// Peek returns the next token but does not advance the Lexer
func (l *Lexer) Peek() (Token, error) {
	if len(l.buffer) > 0 {
		return l.buffer[0], nil
	}

	tok, err := l.Next()
	l.buffer = append(l.buffer, tok)
	return l.buffer[0], err
}

// Next returns the next token and advances the Lexer
func (l *Lexer) Next() (Token, error) {
	if len(l.buffer) > 0 {
		tok := l.buffer[0]
		l.buffer = l.buffer[1:]
		return tok, nil
	}

	return eatToken(l.scanner)
}

// Lex creates a new Lexer struct given a Scanner
func Lex(scanner *Scanner) *Lexer {
	return &Lexer{[]Token{}, scanner}
}

func isEOF(r rune) bool {
	return (r == '\000')
}

func isWhitespace(r rune) bool {
	return (r <= ' ') && (r != '\000')
}

func isLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func isDigit(r rune) bool {
	return ('0' <= r && r <= '9')
}

func eatToken(scanner *Scanner) (Token, error) {
	peek := scanner.Peek()

	switch {
	case isEOF(peek.char):
		return eatEOF(scanner)
	case isWhitespace(peek.char):
		return eatWhitespace(scanner)
	case isLetter(peek.char):
		return eatWordToken(scanner)
	default:
		return Token{}, fmt.Errorf("%s unexpected symbol", peek.loc)
	}
}

func eatEOF(scanner *Scanner) (Token, error) {
	return Token{EOF, "", scanner.Peek().loc}, nil
}

func eatWhitespace(scanner *Scanner) (Token, error) {
	for isWhitespace(scanner.Peek().char) {
		scanner.Next()
	}

	return eatToken(scanner)
}

func eatWordToken(scanner *Scanner) (Token, error) {
	loc := scanner.Peek().loc
	lexeme := ""

	if isLetter(scanner.Peek().char) == false {
		return Token{}, fmt.Errorf("%s expected word", loc)
	}

	for isLetter(scanner.Peek().char) {
		lexeme += string(scanner.Next().char)
	}

	return Token{Ident, lexeme, loc}, nil
}

func eatNumberToken(scanner *Scanner) (Token, error) {
	loc := scanner.Peek().loc
	lexeme := ""

	if isDigit(scanner.Peek().char) == false {
		return Token{}, fmt.Errorf("%s expected number", loc)
	}

	for isDigit(scanner.Peek().char) {
		lexeme += string(scanner.Next().char)
	}

	return Token{Number, lexeme, loc}, nil
}
