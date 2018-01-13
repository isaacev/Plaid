package lang

// tokType classifies the different Tokens
type tokType string

// Token classifications
const (
	tokError    tokType = "Error"
	tokEOF              = "EOF"
	tokPlus             = "+"
	tokDash             = "-"
	tokStar             = "*"
	tokSlash            = "/"
	tokQuestion         = "?"
	tokSemi             = ";"
	tokComma            = ","
	tokParenL           = "("
	tokParenR           = ")"
	tokBraceL           = "{"
	tokBraceR           = "}"
	tokBracketL         = "["
	tokBracketR         = "]"
	tokColon            = ":"
	tokAssign           = ":="
	tokArrow            = "=>"
	tokLT               = "<"
	tokGT               = ">"
	tokLTEquals         = "<="
	tokGTEquals         = ">="
	tokFn               = "fn"
	tokIf               = "if"
	tokLet              = "let"
	tokReturn           = "return"
	tokSelf             = "self"
	tokUse              = "use"
	tokPub              = "pub"
	tokIdent            = "Ident"
	tokNumber           = "Number"
	tokString           = "String"
	tokBoolean          = "Boolean"
)

// token is a basic syntactic unit
type token struct {
	Type   tokType
	Lexeme string
	Loc    Loc
}

// lexer contains methods for generating a sequence of Tokens
type lexer struct {
	Filepath string
	buffer   []token
	scanner  *scanner
}

// peek returns the next token but does not advance the Lexer
func (l *lexer) peek() token {
	if len(l.buffer) > 0 {
		return l.buffer[0]
	}

	tok := l.next()
	l.buffer = append(l.buffer, tok)
	return l.buffer[0]
}

// next returns the next token and advances the Lexer
func (l *lexer) next() token {
	if len(l.buffer) > 0 {
		tok := l.buffer[0]
		l.buffer = l.buffer[1:]
		return tok
	}

	return eatToken(l.scanner)
}

// Lex creates a new Lexer struct given a Scanner
func Lex(filepath string, scn *scanner) *lexer {
	return &lexer{filepath, []token{}, scn}
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

func eatToken(scn *scanner) token {
	peek := scn.peek()

	switch {
	case isEOF(peek.char):
		return eatEOF(scn)
	case isWhitespace(peek.char):
		return eatWhitespace(scn)
	case isOperator(peek.char):
		return eatOperatorToken(scn)
	case isSemicolon(peek.char):
		return eatSemicolonToken(scn)
	case isComma(peek.char):
		return eatCommaToken(scn)
	case isParen(peek.char):
		return eatParenToken(scn)
	case isBrace(peek.char):
		return eatBraceToken(scn)
	case isBracket(peek.char):
		return eatBracketToken(scn)
	case isLetter(peek.char):
		return eatWordToken(scn)
	case isDigit(peek.char):
		return eatNumberToken(scn)
	case isDoubleQuote(peek.char):
		return eatStringToken(scn)
	default:
		return token{tokError, "unexpected symbol", peek.loc}
	}
}

func eatEOF(scn *scanner) token {
	return token{tokEOF, "", scn.peek().loc}
}

func eatWhitespace(scn *scanner) token {
	for isWhitespace(scn.peek().char) {
		scn.next()
	}

	return eatToken(scn)
}

func eatOperatorToken(scn *scanner) token {
	switch scn.peek().char {
	case '+':
		return token{tokPlus, "+", scn.next().loc}
	case '-':
		return token{tokDash, "-", scn.next().loc}
	case '*':
		return token{tokStar, "*", scn.next().loc}
	case '/':
		return token{tokSlash, "/", scn.next().loc}
	case '?':
		return token{tokQuestion, "?", scn.next().loc}
	case ':':
		colon := scn.next()
		tok := token{tokColon, ":", colon.loc}

		if scn.peek().char == '=' {
			scn.next()
			tok = token{tokAssign, ":=", colon.loc}
		}

		return tok
	case '=':
		equals := scn.next()

		if scn.peek().char == '>' {
			scn.next()
			return token{tokArrow, "=>", equals.loc}
		}

		return token{tokError, "expected operator", equals.loc}
	case '<':
		lt := scn.next()

		if scn.peek().char == '=' {
			scn.next()
			return token{tokLTEquals, "<=", lt.loc}
		}

		return token{tokLT, "<", lt.loc}
	case '>':
		gt := scn.next()

		if scn.peek().char == '=' {
			scn.next()
			return token{tokGTEquals, ">=", gt.loc}
		}

		return token{tokGT, ">", gt.loc}
	default:
		return token{tokError, "expected operator", scn.next().loc}
	}
}

func eatSemicolonToken(scn *scanner) token {
	if scn.peek().char != ';' {
		return token{tokError, "expected semicolon", scn.next().loc}
	}

	return token{tokSemi, ";", scn.next().loc}
}

func eatCommaToken(scn *scanner) token {
	if scn.peek().char != ',' {
		return token{tokError, "expected comma", scn.next().loc}
	}

	return token{tokComma, ",", scn.next().loc}
}

func eatParenToken(scn *scanner) token {
	switch scn.peek().char {
	case '(':
		return token{tokParenL, "(", scn.next().loc}
	case ')':
		return token{tokParenR, ")", scn.next().loc}
	default:
		return token{tokError, "expected paren", scn.next().loc}
	}
}

func eatBraceToken(scn *scanner) token {
	switch scn.peek().char {
	case '{':
		return token{tokBraceL, "{", scn.next().loc}
	case '}':
		return token{tokBraceR, "}", scn.next().loc}
	default:
		return token{tokError, "expected brace", scn.next().loc}
	}
}

func eatBracketToken(scn *scanner) token {
	switch scn.peek().char {
	case '[':
		return token{tokBracketL, "[", scn.next().loc}
	case ']':
		return token{tokBracketR, "]", scn.next().loc}
	default:
		return token{tokError, "expected bracket", scn.next().loc}
	}
}

func eatWordToken(scn *scanner) token {
	loc := scn.peek().loc
	lexeme := ""

	if isLetter(scn.peek().char) == false {
		return token{tokError, "expected word", loc}
	}

	for isLetter(scn.peek().char) {
		lexeme += string(scn.next().char)
	}

	switch lexeme {
	case "fn":
		return token{tokFn, "fn", loc}
	case "if":
		return token{tokIf, "if", loc}
	case "let":
		return token{tokLet, "let", loc}
	case "return":
		return token{tokReturn, "return", loc}
	case "self":
		return token{tokSelf, "self", loc}
	case "use":
		return token{tokUse, "use", loc}
	case "pub":
		return token{tokPub, "pub", loc}
	case "true":
		return token{tokBoolean, "true", loc}
	case "false":
		return token{tokBoolean, "false", loc}
	default:
		return token{tokIdent, lexeme, loc}
	}
}

func eatNumberToken(scn *scanner) token {
	loc := scn.peek().loc
	lexeme := ""

	if isDigit(scn.peek().char) == false {
		return token{tokError, "expected number", loc}
	}

	for isDigit(scn.peek().char) {
		lexeme += string(scn.next().char)
	}

	return token{tokNumber, lexeme, loc}
}

func eatStringToken(scn *scanner) token {
	loc := scn.peek().loc
	lexeme := ""

	if isDoubleQuote(scn.peek().char) == false {
		return token{tokError, "expected string", loc}
	}

	lexeme += string(scn.next().char)
	for {
		switch scn.peek().char {
		case '"':
			lexeme += string(scn.next().char)
			return token{tokString, lexeme, loc}
		case '\000':
			fallthrough
		case '\n':
			return token{tokError, "unclosed string", scn.peek().loc}
		default:
			lexeme += string(scn.next().char)
		}
	}
}
