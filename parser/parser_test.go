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
}

func TestParseExpr(t *testing.T) {
	p := makeParser("+a")
	p.registerPrefix(lexer.Plus, parsePrefix)
	p.registerPrefix(lexer.Ident, parseIdent)
	expectExpr(t, p, "(+ a)")

	p = makeParser("+")
	p.registerPrefix(lexer.Plus, parsePrefix)
	p.registerPrefix(lexer.Ident, parseIdent)
	expectExprError(t, p, "unexpected symbol")

	p = makeParser("a + b + c")
	p.registerPostfix(lexer.Plus, parseInfix, Sum)
	p.registerPrefix(lexer.Ident, parseIdent)
	expectExpr(t, p, "(+ (+ a b) c)")

	p = makeParser("a + b * c")
	p.registerPostfix(lexer.Plus, parseInfix, Sum)
	p.registerPostfix(lexer.Star, parseInfix, Product)
	p.registerPrefix(lexer.Ident, parseIdent)
	expectExpr(t, p, "(+ a (* b c))")

	p = makeParser("a +")
	p.registerPostfix(lexer.Plus, parseInfix, Sum)
	p.registerPrefix(lexer.Ident, parseIdent)
	expectExprError(t, p, "unexpected symbol")
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

	if expr, err = parseInfix(parser, left); err != nil {
		t.Errorf("Expected no errors, got %v\n", err)
	} else if expr.String() != "(+ a b)" {
		t.Errorf("Expected (+ a b), got %s\n", expr)
	}

	parser = makeParser("a +")
	parser.registerPostfix(lexer.Plus, parseInfix, Sum)
	parser.registerPrefix(lexer.Ident, parseIdent)

	if left, err = parseIdent(parser); err != nil {
		t.Errorf("Expected no errors, got %v\n", err)
	}

	if expr, err = parseInfix(parser, left); err == nil {
		t.Errorf("Expected an error, got %s\n", expr)
	} else if err.Error() != "unexpected symbol" {
		t.Errorf("Expected '%s', got %v\n", "unexpected symbol", err)
	}
}

func TestParsePrefix(t *testing.T) {
	parser := makeParser("+a")
	parser.registerPrefix(lexer.Plus, parsePrefix)
	parser.registerPrefix(lexer.Ident, parseIdent)

	if expr, err := parsePrefix(parser); err != nil {
		t.Errorf("Expected no errors, got %v\n", err)
	} else if expr.String() != "(+ a)" {
		t.Errorf("Expected '(+ a)', got '%s'\n", expr.String())
	}

	parser = makeParser("+")
	parser.registerPrefix(lexer.Plus, parsePrefix)
	parser.registerPrefix(lexer.Ident, parseIdent)

	if expr, err := parsePrefix(parser); err == nil {
		t.Errorf("Expected an error, got %s\n", expr)
	} else if err.Error() != "unexpected symbol" {
		t.Errorf("Expected '%s', got %v\n", "unexpected symbol", err)
	}
}

func TestParseIdent(t *testing.T) {
	parser := makeParser("abc")

	if expr, err := parseIdent(parser); err != nil {
		t.Errorf("Expected no errors, got %v\n", err)
	} else if expr.String() != "abc" {
		t.Errorf("Expected 'abc', got '%s'\n", expr.String())
	}

	parser = makeParser("123")

	if expr, err := parseIdent(parser); err == nil {
		t.Errorf("Expected an error, got %s\n", expr)
	} else if err.Error() != "expected identifier" {
		t.Errorf("Expected '%s', got %v\n", "expected identifier", err)
	}
}

func TestParseNumber(t *testing.T) {
	parser := makeParser("123")

	if expr, err := parseNumber(parser); err != nil {
		t.Errorf("Expected no errors, got %v\n", err)
	} else if expr.String() != "123" {
		t.Errorf("Expected '123', got '%s'\n", expr.String())
	}

	parser = makeParser("abc")

	if expr, err := parseNumber(parser); err == nil {
		t.Errorf("Expected an error, got %s\n", expr)
	} else if err.Error() != "expected number literal" {
		t.Errorf("Expected '%s', got %v\n", "expected number literal", err)
	}

	loc := lexer.Loc{Line: 1, Col: 1}
	expr, err := evalNumber(lexer.Token{Type: lexer.Number, Lexeme: "abc", Loc: loc})
	if err == nil {
		t.Errorf("Expected an error, got %s\n", expr)
	} else if err.Error() != "malformed number literal" {
		t.Errorf("Expected '%s', got %v\n", "malformed number literal", err)
	}
}

func makeParser(source string) *Parser {
	return &Parser{
		lexer.Lex(lexer.Scan(source)),
		make(map[lexer.Type]Precedence),
		make(map[lexer.Type]PrefixParseFunc),
		make(map[lexer.Type]PostfixParseFunc),
	}
}

func expectExpr(t *testing.T, p *Parser, ast string) {
	if expr, err := parseExpr(p, Lowest); err != nil {
		t.Errorf("Expected no errors, got %v\n", err)
	} else {
		expectAST(t, ast, expr)
	}
}

func expectExprError(t *testing.T, p *Parser, msg string) {
	if expr, err := parseExpr(p, Lowest); err == nil {
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
