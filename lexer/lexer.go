package lexer

// Type classifies the different Tokens
type Type string

// Token classifications
const (
	Error  Type = "Error"
	EOF         = "EOF"
	Plus        = "+"
	Dash        = "-"
	Star        = "*"
	Slash       = "/"
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

// Type returns the Token classification
func (t Token) Type() Type {
	return t.typ
}

// Lexeme returns the Token lexeme string
func (t Token) Lexeme() string {
	return t.lexeme
}

// Loc returns the Token Loc struct
func (t Token) Loc() Loc {
	return t.loc
}

// Lexer contains methods for generating a sequence of Tokens
type Lexer struct {
	buffer  []Token
	scanner *Scanner
}

// Peek returns the next token but does not advance the Lexer
func (l *Lexer) Peek() Token {
	if len(l.buffer) > 0 {
		return l.buffer[0]
	}

	tok := l.Next()
	l.buffer = append(l.buffer, tok)
	return l.buffer[0]
}

// Next returns the next token and advances the Lexer
func (l *Lexer) Next() Token {
	if len(l.buffer) > 0 {
		tok := l.buffer[0]
		l.buffer = l.buffer[1:]
		return tok
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

func isOperator(r rune) bool {
	switch r {
	case '+':
		fallthrough
	case '-':
		fallthrough
	case '*':
		fallthrough
	case '/':
		return true
	default:
		return false
	}
}

func isLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func isDigit(r rune) bool {
	return ('0' <= r && r <= '9')
}

func eatToken(scanner *Scanner) Token {
	peek := scanner.Peek()

	switch {
	case isEOF(peek.char):
		return eatEOF(scanner)
	case isWhitespace(peek.char):
		return eatWhitespace(scanner)
	case isOperator(peek.char):
		return eatOperatorToken(scanner)
	case isLetter(peek.char):
		return eatWordToken(scanner)
	case isDigit(peek.char):
		return eatNumberToken(scanner)
	default:
		return Token{Error, "unexpected symbol", peek.loc}
	}
}

func eatEOF(scanner *Scanner) Token {
	return Token{EOF, "", scanner.Peek().loc}
}

func eatWhitespace(scanner *Scanner) Token {
	for isWhitespace(scanner.Peek().char) {
		scanner.Next()
	}

	return eatToken(scanner)
}

func eatOperatorToken(scanner *Scanner) Token {
	switch scanner.Peek().char {
	case '+':
		return Token{Plus, "+", scanner.Next().loc}
	case '-':
		return Token{Dash, "-", scanner.Next().loc}
	case '*':
		return Token{Star, "*", scanner.Next().loc}
	case '/':
		return Token{Slash, "/", scanner.Next().loc}
	default:
		return Token{Error, "expected operator", scanner.Next().loc}
	}
}

func eatWordToken(scanner *Scanner) Token {
	loc := scanner.Peek().loc
	lexeme := ""

	if isLetter(scanner.Peek().char) == false {
		return Token{Error, "expected word", loc}
	}

	for isLetter(scanner.Peek().char) {
		lexeme += string(scanner.Next().char)
	}

	return Token{Ident, lexeme, loc}
}

func eatNumberToken(scanner *Scanner) Token {
	loc := scanner.Peek().loc
	lexeme := ""

	if isDigit(scanner.Peek().char) == false {
		return Token{Error, "expected number", loc}
	}

	for isDigit(scanner.Peek().char) {
		lexeme += string(scanner.Next().char)
	}

	return Token{Number, lexeme, loc}
}
