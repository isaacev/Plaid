package parser

import (
	"fmt"
	"plaid/lexer"
	"strconv"
)

// Precedence describes the relative binding powers of different operators
type Precedence int

// The staticly defined precedence levels
const (
	Lowest Precedence = iota
	Sum
	Product
	Prefix
	Postfix
)

// PrefixParseFunc describes the parsing function for any construct where the
// binding operator comes before the expression it binds to.
type PrefixParseFunc func(p *Parser) (Expr, error)

// PostfixParseFunc describes the parsing function for any construct where the
// binding operator comes after the expression it binds to.
type PostfixParseFunc func(p *Parser, left Expr) (Expr, error)

// Parser contains methods for generating an abstract syntax tree from a
// sequence of Tokens
type Parser struct {
	lexer             *lexer.Lexer
	precedenceTable   map[lexer.Type]Precedence
	prefixParseFuncs  map[lexer.Type]PrefixParseFunc
	postfixParseFuncs map[lexer.Type]PostfixParseFunc
}

func (p *Parser) peekTokenIsNot(typ lexer.Type) bool {
	return (p.lexer.Peek().Type != typ)
}

func (p *Parser) registerPrecedence(typ lexer.Type, level Precedence) {
	p.precedenceTable[typ] = level
}

func (p *Parser) registerPrefix(typ lexer.Type, fn PrefixParseFunc) {
	p.prefixParseFuncs[typ] = fn
}

func (p *Parser) registerPostfix(typ lexer.Type, fn PostfixParseFunc, level Precedence) {
	p.registerPrecedence(typ, level)
	p.postfixParseFuncs[typ] = fn
}

func (p *Parser) peekPrecedence() Precedence {
	prec, exists := p.precedenceTable[p.lexer.Peek().Type]

	if exists {
		return prec
	}

	return Lowest
}

// Parse initializers a parser and defines the grammar precedence levels
func Parse(l *lexer.Lexer) *Parser {
	parser := &Parser{
		l,
		make(map[lexer.Type]Precedence),
		make(map[lexer.Type]PrefixParseFunc),
		make(map[lexer.Type]PostfixParseFunc),
	}

	parser.registerPrefix(lexer.ParenL, parseGroup)
	parser.registerPrefix(lexer.Plus, parsePrefix)
	parser.registerPrefix(lexer.Dash, parsePrefix)
	parser.registerPrefix(lexer.Ident, parseIdent)
	parser.registerPrefix(lexer.Number, parseNumber)

	parser.registerPostfix(lexer.Plus, parseInfix, Sum)
	parser.registerPostfix(lexer.Dash, parseInfix, Sum)
	parser.registerPostfix(lexer.Star, parseInfix, Product)
	parser.registerPostfix(lexer.Slash, parseInfix, Product)

	return parser
}

func parseStmt(p *Parser) (Stmt, error) {
	switch p.lexer.Peek().Type {
	case lexer.Let:
		return parseDeclarationStmt(p)
	default:
		return nil, fmt.Errorf("expected start of statement")
	}
}

func parseDeclarationStmt(p *Parser) (Stmt, error) {
	if p.peekTokenIsNot(lexer.Let) {
		return nil, fmt.Errorf("expected LET keyword")
	}

	tok := p.lexer.Next()
	var expr Expr
	var err error
	if expr, err = parseIdent(p); err != nil {
		return nil, err
	}

	name := expr.(IdentExpr)

	if p.peekTokenIsNot(lexer.Assign) {
		return nil, fmt.Errorf("expected :=")
	}

	p.lexer.Next()

	expr, err = parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	if p.peekTokenIsNot(lexer.Semi) {
		return nil, fmt.Errorf("expected semicolon")
	}

	p.lexer.Next()
	return DeclarationStmt{tok, name, expr}, nil
}

func parseExpr(p *Parser, level Precedence) (Expr, error) {
	prefix, exists := p.prefixParseFuncs[p.lexer.Peek().Type]
	if exists == false {
		return nil, fmt.Errorf("unexpected symbol")
	}

	left, err := prefix(p)
	if err != nil {
		return nil, err
	}

	for p.peekTokenIsNot(lexer.EOF) && level < p.peekPrecedence() {
		infix := p.postfixParseFuncs[p.lexer.Peek().Type]
		left, err = infix(p, left)

		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func parseInfix(p *Parser, left Expr) (Expr, error) {
	level := p.peekPrecedence()
	tok := p.lexer.Next()
	oper := tok.Lexeme
	right, err := parseExpr(p, level)
	if err != nil {
		return nil, err
	}

	return BinaryExpr{oper, tok, left, right}, nil
}

func parsePostfix(p *Parser, left Expr) (Expr, error) {
	tok := p.lexer.Next()
	oper := tok.Lexeme

	return UnaryExpr{oper, tok, left}, nil
}

func parsePrefix(p *Parser) (Expr, error) {
	tok := p.lexer.Next()
	oper := tok.Lexeme
	right, err := parseExpr(p, Prefix)
	if err != nil {
		return nil, err
	}

	return UnaryExpr{oper, tok, right}, nil
}

func parseGroup(p *Parser) (Expr, error) {
	if p.peekTokenIsNot(lexer.ParenL) {
		return nil, fmt.Errorf("expected left paren")
	}

	p.lexer.Next()
	expr, err := parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	if p.peekTokenIsNot(lexer.ParenR) {
		return nil, fmt.Errorf("expected right paren")
	}

	p.lexer.Next()
	return expr, nil
}

func parseIdent(p *Parser) (Expr, error) {
	if p.peekTokenIsNot(lexer.Ident) {
		return nil, fmt.Errorf("expected identifier")
	}

	tok := p.lexer.Next()
	return IdentExpr{tok, tok.Lexeme}, nil
}

func parseNumber(p *Parser) (Expr, error) {
	if p.peekTokenIsNot(lexer.Number) {
		return nil, fmt.Errorf("expected number literal")
	}

	tok := p.lexer.Next()
	return evalNumber(tok)
}

func evalNumber(tok lexer.Token) (NumberExpr, error) {
	val, err := strconv.ParseUint(tok.Lexeme, 10, 64)
	if err != nil {
		return NumberExpr{}, fmt.Errorf("malformed number literal")
	}

	return NumberExpr{tok, int(val)}, nil
}
