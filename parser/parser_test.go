package parser

import (
	"plaid/lexer"
	"testing"
)

func TestPeekTokenIsNot(t *testing.T) {
	parser := makeParser("abc")

	if parser.peekTokenIsNot(lexer.Ident) == true {
		t.Errorf("Expected Parser.peekTokenIsNot %t, got %t\n", false, true)
	}
}

func TestRegisterPrecedence(t *testing.T) {
	parser := makeParser("")
	parser.registerPrecedence(lexer.Ident, Sum)

	got := parser.precedenceTable[lexer.Ident]
	if got != Sum {
		t.Errorf("Expected %v, got %v\n", Sum, got)
	}
}

func TestRegisterPrefix(t *testing.T) {
	parser := makeParser("")
	parser.registerPrefix(lexer.Ident, parseIdent)

	if _, exists := parser.prefixParseFuncs[lexer.Ident]; exists == false {
		t.Error("Expected prefix parse function, got nothing")
	}
}

func TestRegisterPostfix(t *testing.T) {
	parser := makeParser("")
	parser.registerPostfix(lexer.Plus, parseInfix, Sum)

	if _, exists := parser.postfixParseFuncs[lexer.Plus]; exists == false {
		t.Error("Expected postfix parse function, got nothing")
	}

	level, exists := parser.precedenceTable[lexer.Plus]
	if (exists == false) || (level != Sum) {
		t.Errorf("Expected Plus precedence to be %v, got %v\n", Sum, level)
	}
}

func TestPeekPrecedence(t *testing.T) {
	parser := makeParser("+*")
	parser.registerPrecedence(lexer.Plus, Sum)

	level := parser.peekPrecedence()
	if level != Sum {
		t.Errorf("Expected Plus precedence to be %v, got %v\n", Sum, level)
	}

	parser.lexer.Next()

	level = parser.peekPrecedence()
	if level != Lowest {
		t.Errorf("Expected Star precedence to be %v, got %v\n", Lowest, level)
	}
}

func TestParseInitializer(t *testing.T) {
	expectParse := func(source string, ast string) {
		if expr, err := Parse(lexer.Lex(lexer.Scan(source))); err != nil {
			t.Errorf("Expected no errors, got '%s'\n", err)
		} else {
			expectAST(t, ast, expr)
		}
	}

	expectParse("a + b", "(+ a b)")
	expectParse("2 + 2", "(+ 2 2)")
	expectParse("2 + xyz", "(+ 2 xyz)")
	expectParse("a - b", "(- a b)")
	expectParse("a + + b", "(+ a (+ b))")
	expectParse("+a + + b", "(+ (+ a) (+ b))")
	expectParse("a + b + c", "(+ (+ a b) c)")
	expectParse("a * b + c", "(+ (* a b) c)")
	expectParse("a + b * c", "(+ a (* b c))")
	expectParse("(a + b) * c", "(* (+ a b) c)")
}

func TestParseExpr(t *testing.T) {
	p := makeParser("+a")
	p.registerPrefix(lexer.Plus, parsePrefix)
	p.registerPrefix(lexer.Ident, parseIdent)
	expr, err := parseExpr(p, Lowest)
	expectNoErrors(t, "(+ a)", expr, err)

	p = makeParser("+")
	p.registerPrefix(lexer.Plus, parsePrefix)
	p.registerPrefix(lexer.Ident, parseIdent)
	expr, err = parseExpr(p, Lowest)
	expectAnError(t, "unexpected symbol", expr, err)

	p = makeParser("a + b + c")
	p.registerPostfix(lexer.Plus, parseInfix, Sum)
	p.registerPrefix(lexer.Ident, parseIdent)
	expr, err = parseExpr(p, Lowest)
	expectNoErrors(t, "(+ (+ a b) c)", expr, err)

	p = makeParser("a + b * c")
	p.registerPostfix(lexer.Plus, parseInfix, Sum)
	p.registerPostfix(lexer.Star, parseInfix, Product)
	p.registerPrefix(lexer.Ident, parseIdent)
	expr, err = parseExpr(p, Lowest)
	expectNoErrors(t, "(+ a (* b c))", expr, err)

	p = makeParser("a +")
	p.registerPostfix(lexer.Plus, parseInfix, Sum)
	p.registerPrefix(lexer.Ident, parseIdent)
	expr, err = parseExpr(p, Lowest)
	expectAnError(t, "unexpected symbol", expr, err)
}

func TestParseInfix(t *testing.T) {
	parser := makeParser("a + b")
	parser.registerPostfix(lexer.Plus, parseInfix, Sum)
	parser.registerPrefix(lexer.Ident, parseIdent)

	var left Expr
	var expr Expr
	var err error

	if left, err = parseIdent(parser); err != nil {
		t.Errorf("Expected no errors, got %v\n", err)
	}

	expr, err = parseInfix(parser, left)
	expectNoErrors(t, "(+ a b)", expr, err)

	parser = makeParser("a +")
	parser.registerPostfix(lexer.Plus, parseInfix, Sum)
	parser.registerPrefix(lexer.Ident, parseIdent)

	if left, err = parseIdent(parser); err != nil {
		t.Errorf("Expected no errors, got %v\n", err)
	}

	expr, err = parseInfix(parser, left)
	expectAnError(t, "unexpected symbol", expr, err)
}

func TestParsePostfix(t *testing.T) {
	parser := makeParser("a+")
	parser.registerPostfix(lexer.Plus, parsePostfix, Postfix)
	parser.registerPrefix(lexer.Ident, parseIdent)

	var left Expr
	var expr Expr
	var err error

	if left, err = parseIdent(parser); err != nil {
		t.Errorf("Expected no errors, got %v\n", err)
	}

	expr, err = parsePostfix(parser, left)
	expectNoErrors(t, "(+ a)", expr, err)
}

func TestParsePrefix(t *testing.T) {
	parser := makeParser("+a")
	parser.registerPrefix(lexer.Plus, parsePrefix)
	parser.registerPrefix(lexer.Ident, parseIdent)

	expr, err := parsePrefix(parser)
	expectNoErrors(t, "(+ a)", expr, err)

	parser = makeParser("+")
	parser.registerPrefix(lexer.Plus, parsePrefix)
	parser.registerPrefix(lexer.Ident, parseIdent)

	expr, err = parsePrefix(parser)
	expectAnError(t, "unexpected symbol", expr, err)
}

func TestParseGroup(t *testing.T) {
	parser := makeParser("(a)")
	parser.registerPrefix(lexer.ParenL, parsePrefix)
	parser.registerPrefix(lexer.Ident, parseIdent)

	expr, err := parseGroup(parser)
	expectNoErrors(t, "a", expr, err)

	parser = makeParser("a)")
	parser.registerPrefix(lexer.ParenL, parsePrefix)
	parser.registerPrefix(lexer.Ident, parseIdent)

	expr, err = parseGroup(parser)
	expectAnError(t, "expected left paren", expr, err)

	parser = makeParser("(")
	parser.registerPrefix(lexer.ParenL, parsePrefix)
	parser.registerPrefix(lexer.Ident, parseIdent)

	expr, err = parseGroup(parser)
	expectAnError(t, "unexpected symbol", expr, err)

	parser = makeParser("(a")
	parser.registerPrefix(lexer.ParenL, parsePrefix)
	parser.registerPrefix(lexer.Ident, parseIdent)

	expr, err = parseGroup(parser)
	expectAnError(t, "expected right paren", expr, err)
}

func TestParseIdent(t *testing.T) {
	parser := makeParser("abc")

	expr, err := parseIdent(parser)
	expectNoErrors(t, "abc", expr, err)

	parser = makeParser("123")

	expr, err = parseIdent(parser)
	expectAnError(t, "expected identifier", expr, err)
}

func TestParseNumber(t *testing.T) {
	parser := makeParser("123")

	expr, err := parseNumber(parser)
	expectNoErrors(t, "123", expr, err)

	parser = makeParser("abc")

	expr, err = parseNumber(parser)
	expectAnError(t, "expected number literal", expr, err)

	loc := lexer.Loc{Line: 1, Col: 1}
	expr, err = evalNumber(lexer.Token{Type: lexer.Number, Lexeme: "abc", Loc: loc})
	expectAnError(t, "malformed number literal", expr, err)
}

func makeParser(source string) *Parser {
	return &Parser{
		lexer.Lex(lexer.Scan(source)),
		make(map[lexer.Type]Precedence),
		make(map[lexer.Type]PrefixParseFunc),
		make(map[lexer.Type]PostfixParseFunc),
	}
}

func expectNoErrors(t *testing.T, ast string, expr Expr, err error) {
	if err != nil {
		t.Errorf("Expected no errors, got '%s'\n", err)
	} else {
		expectAST(t, ast, expr)
	}
}

func expectAnError(t *testing.T, msg string, expr Expr, err error) {
	if err == nil {
		t.Errorf("Expected an error, got %s\n", expr)
	} else if err.Error() != msg {
		t.Errorf("Expected '%s', got '%s'\n", msg, err)
	}
}

func expectAST(t *testing.T, ast string, got Expr) {
	if ast != got.String() {
		t.Errorf("Expected '%s', got '%s'\n", ast, got)
	}
}
