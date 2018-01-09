package lexer

// TokenType classifies the different Tokens
type TokenType string

// Token classifications
const (
	TokError    TokenType = "Error"
	TokEOF                = "EOF"
	TokPlus               = "+"
	TokDash               = "-"
	TokStar               = "*"
	TokSlash              = "/"
	TokQuestion           = "?"
	TokSemi               = ";"
	TokComma              = ","
	TokParenL             = "("
	TokParenR             = ")"
	TokBraceL             = "{"
	TokBraceR             = "}"
	TokBracketL           = "["
	TokBracketR           = "]"
	TokColon              = ":"
	TokAssign             = ":="
	TokArrow              = "=>"
	TokLT                 = "<"
	TokGT                 = ">"
	TokLTEquals           = "<="
	TokGTEquals           = ">="
	TokFn                 = "fn"
	TokIf                 = "if"
	TokLet                = "let"
	TokReturn             = "return"
	TokSelf               = "self"
	TokUse                = "use"
	TokPub                = "pub"
	TokIdent              = "Ident"
	TokNumber             = "Number"
	TokString             = "String"
	TokBoolean            = "Boolean"
)

// Token is a basic syntactic unit
type Token struct {
	Type   TokenType
	Lexeme string
	Loc    Loc
}

// Lexer contains methods for generating a sequence of Tokens
type Lexer struct {
	Filepath string
	buffer   []Token
	scanner  *Scanner
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
func Lex(filepath string, scanner *Scanner) *Lexer {
	return &Lexer{filepath, []Token{}, scanner}
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
	case '<':
		return true
	case '>':
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
		return Token{TokError, "unexpected symbol", peek.loc}
	}
}

func eatEOF(scanner *Scanner) Token {
	return Token{TokEOF, "", scanner.Peek().loc}
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
		return Token{TokPlus, "+", scanner.Next().loc}
	case '-':
		return Token{TokDash, "-", scanner.Next().loc}
	case '*':
		return Token{TokStar, "*", scanner.Next().loc}
	case '/':
		return Token{TokSlash, "/", scanner.Next().loc}
	case '?':
		return Token{TokQuestion, "?", scanner.Next().loc}
	case ':':
		colon := scanner.Next()
		tok := Token{TokColon, ":", colon.loc}

		if scanner.Peek().char == '=' {
			scanner.Next()
			tok = Token{TokAssign, ":=", colon.loc}
		}

		return tok
	case '=':
		equals := scanner.Next()

		if scanner.Peek().char == '>' {
			scanner.Next()
			return Token{TokArrow, "=>", equals.loc}
		}

		return Token{TokError, "expected operator", equals.loc}
	case '<':
		lt := scanner.Next()

		if scanner.Peek().char == '=' {
			scanner.Next()
			return Token{TokLTEquals, "<=", lt.loc}
		}

		return Token{TokLT, "<", lt.loc}
	case '>':
		gt := scanner.Next()

		if scanner.Peek().char == '=' {
			scanner.Next()
			return Token{TokGTEquals, ">=", gt.loc}
		}

		return Token{TokGT, ">", gt.loc}
	default:
		return Token{TokError, "expected operator", scanner.Next().loc}
	}
}

func eatSemicolonToken(scanner *Scanner) Token {
	if scanner.Peek().char != ';' {
		return Token{TokError, "expected semicolon", scanner.Next().loc}
	}

	return Token{TokSemi, ";", scanner.Next().loc}
}

func eatCommaToken(scanner *Scanner) Token {
	if scanner.Peek().char != ',' {
		return Token{TokError, "expected comma", scanner.Next().loc}
	}

	return Token{TokComma, ",", scanner.Next().loc}
}

func eatParenToken(scanner *Scanner) Token {
	switch scanner.Peek().char {
	case '(':
		return Token{TokParenL, "(", scanner.Next().loc}
	case ')':
		return Token{TokParenR, ")", scanner.Next().loc}
	default:
		return Token{TokError, "expected paren", scanner.Next().loc}
	}
}

func eatBraceToken(scanner *Scanner) Token {
	switch scanner.Peek().char {
	case '{':
		return Token{TokBraceL, "{", scanner.Next().loc}
	case '}':
		return Token{TokBraceR, "}", scanner.Next().loc}
	default:
		return Token{TokError, "expected brace", scanner.Next().loc}
	}
}

func eatBracketToken(scanner *Scanner) Token {
	switch scanner.Peek().char {
	case '[':
		return Token{TokBracketL, "[", scanner.Next().loc}
	case ']':
		return Token{TokBracketR, "]", scanner.Next().loc}
	default:
		return Token{TokError, "expected bracket", scanner.Next().loc}
	}
}

func eatWordToken(scanner *Scanner) Token {
	loc := scanner.Peek().loc
	lexeme := ""

	if isLetter(scanner.Peek().char) == false {
		return Token{TokError, "expected word", loc}
	}

	for isLetter(scanner.Peek().char) {
		lexeme += string(scanner.Next().char)
	}

	switch lexeme {
	case "fn":
		return Token{TokFn, "fn", loc}
	case "if":
		return Token{TokIf, "if", loc}
	case "let":
		return Token{TokLet, "let", loc}
	case "return":
		return Token{TokReturn, "return", loc}
	case "self":
		return Token{TokSelf, "self", loc}
	case "use":
		return Token{TokUse, "use", loc}
	case "pub":
		return Token{TokPub, "pub", loc}
	case "true":
		return Token{TokBoolean, "true", loc}
	case "false":
		return Token{TokBoolean, "false", loc}
	default:
		return Token{TokIdent, lexeme, loc}
	}
}

func eatNumberToken(scanner *Scanner) Token {
	loc := scanner.Peek().loc
	lexeme := ""

	if isDigit(scanner.Peek().char) == false {
		return Token{TokError, "expected number", loc}
	}

	for isDigit(scanner.Peek().char) {
		lexeme += string(scanner.Next().char)
	}

	return Token{TokNumber, lexeme, loc}
}

func eatStringToken(scanner *Scanner) Token {
	loc := scanner.Peek().loc
	lexeme := ""

	if isDoubleQuote(scanner.Peek().char) == false {
		return Token{TokError, "expected string", loc}
	}

	lexeme += string(scanner.Next().char)
	for {
		switch scanner.Peek().char {
		case '"':
			lexeme += string(scanner.Next().char)
			return Token{TokString, lexeme, loc}
		case '\000':
			fallthrough
		case '\n':
			return Token{TokError, "unclosed string", scanner.Peek().loc}
		default:
			lexeme += string(scanner.Next().char)
		}
	}
}
