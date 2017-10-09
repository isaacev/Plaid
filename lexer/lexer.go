package lexer

// Type classifies the different Tokens
type Type string

// Token classifications
const (
	Error    Type = "Error"
	EOF           = "EOF"
	Plus          = "+"
	Dash          = "-"
	Star          = "*"
	Slash         = "/"
	Question      = "?"
	Semi          = ";"
	Comma         = ","
	ParenL        = "("
	ParenR        = ")"
	BraceL        = "{"
	BraceR        = "}"
	BracketL      = "["
	BracketR      = "]"
	Colon         = ":"
	Assign        = ":="
	Arrow         = "=>"
	Fn            = "fn"
	If            = "if"
	Let           = "let"
	Return        = "return"
	Ident         = "Ident"
	Number        = "Number"
	String        = "String"
	Boolean       = "Boolean"
)

// Token is a basic syntactic unit
type Token struct {
	Type   Type
	Lexeme string
	Loc    Loc
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
	case '?':
		return true
	case ':':
		return true
	case '=':
		return true
	default:
		return false
	}
}

func isSemicolon(r rune) bool {
	return (r == ';')
}

func isComma(r rune) bool {
	return (r == ',')
}

func isParen(r rune) bool {
	return (r == '(') || (r == ')')
}

func isBrace(r rune) bool {
	return (r == '{') || (r == '}')
}

func isBracket(r rune) bool {
	return (r == '[') || (r == ']')
}

func isLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func isDigit(r rune) bool {
	return ('0' <= r && r <= '9')
}

func isDoubleQuote(r rune) bool {
	return (r == '"')
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
	case isSemicolon(peek.char):
		return eatSemicolonToken(scanner)
	case isComma(peek.char):
		return eatCommaToken(scanner)
	case isParen(peek.char):
		return eatParenToken(scanner)
	case isBrace(peek.char):
		return eatBraceToken(scanner)
	case isBracket(peek.char):
		return eatBracketToken(scanner)
	case isLetter(peek.char):
		return eatWordToken(scanner)
	case isDigit(peek.char):
		return eatNumberToken(scanner)
	case isDoubleQuote(peek.char):
		return eatStringToken(scanner)
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
	case '?':
		return Token{Question, "?", scanner.Next().loc}
	case ':':
		colon := scanner.Next()
		tok := Token{Colon, ":", colon.loc}

		if scanner.Peek().char == '=' {
			scanner.Next()
			tok = Token{Assign, ":=", colon.loc}
		}

		return tok
	case '=':
		equals := scanner.Next()

		if scanner.Peek().char == '>' {
			scanner.Next()
			return Token{Arrow, "=>", equals.loc}
		}

		return Token{Error, "expected operator", equals.loc}
	default:
		return Token{Error, "expected operator", scanner.Next().loc}
	}
}

func eatSemicolonToken(scanner *Scanner) Token {
	if scanner.Peek().char != ';' {
		return Token{Error, "expected semicolon", scanner.Next().loc}
	}

	return Token{Semi, ";", scanner.Next().loc}
}

func eatCommaToken(scanner *Scanner) Token {
	if scanner.Peek().char != ',' {
		return Token{Error, "expected comma", scanner.Next().loc}
	}

	return Token{Comma, ",", scanner.Next().loc}
}

func eatParenToken(scanner *Scanner) Token {
	switch scanner.Peek().char {
	case '(':
		return Token{ParenL, "(", scanner.Next().loc}
	case ')':
		return Token{ParenR, ")", scanner.Next().loc}
	default:
		return Token{Error, "expected paren", scanner.Next().loc}
	}
}

func eatBraceToken(scanner *Scanner) Token {
	switch scanner.Peek().char {
	case '{':
		return Token{BraceL, "{", scanner.Next().loc}
	case '}':
		return Token{BraceR, "}", scanner.Next().loc}
	default:
		return Token{Error, "expected brace", scanner.Next().loc}
	}
}

func eatBracketToken(scanner *Scanner) Token {
	switch scanner.Peek().char {
	case '[':
		return Token{BracketL, "[", scanner.Next().loc}
	case ']':
		return Token{BracketR, "]", scanner.Next().loc}
	default:
		return Token{Error, "expected bracket", scanner.Next().loc}
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

	switch lexeme {
	case "fn":
		return Token{Fn, "fn", loc}
	case "if":
		return Token{If, "if", loc}
	case "let":
		return Token{Let, "let", loc}
	case "return":
		return Token{Return, "return", loc}
	case "true":
		return Token{Boolean, "true", loc}
	case "false":
		return Token{Boolean, "false", loc}
	default:
		return Token{Ident, lexeme, loc}
	}
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

func eatStringToken(scanner *Scanner) Token {
	loc := scanner.Peek().loc
	lexeme := ""

	if isDoubleQuote(scanner.Peek().char) == false {
		return Token{Error, "expected string", loc}
	}

	lexeme += string(scanner.Next().char)
	for {
		switch scanner.Peek().char {
		case '"':
			lexeme += string(scanner.Next().char)
			return Token{String, lexeme, loc}
		case '\000':
			fallthrough
		case '\n':
			return Token{Error, "unclosed string", scanner.Peek().loc}
		default:
			lexeme += string(scanner.Next().char)
		}
	}
}
