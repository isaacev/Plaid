package parser

import (
	"fmt"
	"plaid/lexer"
	"strconv"
)

// SyntaxError combines a source code location with the resulting error message
type SyntaxError struct {
	loc lexer.Loc
	msg string
}

func (se SyntaxError) Error() string {
	return fmt.Sprintf("%s %s", se.loc, se.msg)
}

func makeSyntaxError(tok lexer.Token, msg string, deference bool) error {
	if tok.Type == lexer.Error && deference {
		msg = tok.Lexeme
	}

	return SyntaxError{tok.Loc, msg}
}

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

func (p *Parser) peekTokenIsNot(first lexer.Type, rest ...lexer.Type) bool {
	peek := p.lexer.Peek().Type

	if first == peek {
		return false
	}

	for _, other := range rest {
		if other == peek {
			return false
		}
	}

	return true
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
func Parse(source string) (Program, error) {
	p := makeParser(source)
	loadGrammar(p)
	return parseProgram(p)
}

func makeParser(source string) *Parser {
	s := lexer.Scan(source)
	l := lexer.Lex(s)
	p := &Parser{
		l,
		make(map[lexer.Type]Precedence),
		make(map[lexer.Type]PrefixParseFunc),
		make(map[lexer.Type]PostfixParseFunc),
	}

	return p
}

func loadGrammar(p *Parser) {
	p.registerPrefix(lexer.Fn, parseFunction)
	p.registerPrefix(lexer.ParenL, parseGroup)
	p.registerPrefix(lexer.Plus, parsePrefix)
	p.registerPrefix(lexer.Dash, parsePrefix)
	p.registerPrefix(lexer.Ident, parseIdent)
	p.registerPrefix(lexer.Number, parseNumber)

	p.registerPostfix(lexer.Plus, parseInfix, Sum)
	p.registerPostfix(lexer.Dash, parseInfix, Sum)
	p.registerPostfix(lexer.Star, parseInfix, Product)
	p.registerPostfix(lexer.Slash, parseInfix, Product)
}

func parseProgram(p *Parser) (Program, error) {
	stmts := []Stmt{}

	for p.peekTokenIsNot(lexer.Error, lexer.EOF) {
		stmt, err := parseStmt(p)
		if err != nil {
			return Program{}, err
		}

		stmts = append(stmts, stmt)
	}

	return Program{stmts}, nil
}

func parseStmt(p *Parser) (Stmt, error) {
	switch p.lexer.Peek().Type {
	case lexer.Let:
		return parseDeclarationStmt(p)
	default:
		return nil, makeSyntaxError(p.lexer.Peek(), "expected start of statement", false)
	}
}

func parseStmtBlock(p *Parser) (StmtBlock, error) {
	if p.peekTokenIsNot(lexer.BraceL) {
		return StmtBlock{}, makeSyntaxError(p.lexer.Peek(), "expected left brace", false)
	}

	left := p.lexer.Next()
	stmts := []Stmt{}
	for p.peekTokenIsNot(lexer.BraceR, lexer.EOF, lexer.Error) {
		stmt, err := parseStmt(p)
		if err != nil {
			return StmtBlock{}, err
		}

		stmts = append(stmts, stmt)
	}

	if p.peekTokenIsNot(lexer.BraceR) {
		return StmtBlock{}, makeSyntaxError(p.lexer.Peek(), "expected right brace", false)
	}

	right := p.lexer.Next()
	return StmtBlock{left, stmts, right}, nil
}

func parseDeclarationStmt(p *Parser) (Stmt, error) {
	if p.peekTokenIsNot(lexer.Let) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected LET keyword", false)
	}

	tok := p.lexer.Next()
	var expr Expr
	var err error
	if expr, err = parseIdent(p); err != nil {
		return nil, err
	}

	name := expr.(IdentExpr)

	if p.peekTokenIsNot(lexer.Assign) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected :=", false)
	}

	p.lexer.Next()

	expr, err = parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	if p.peekTokenIsNot(lexer.Semi) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected semicolon", false)
	}

	p.lexer.Next()
	return DeclarationStmt{tok, name, expr}, nil
}

func parseTypeSig(p *Parser) (TypeSig, error) {
	var child TypeSig
	var err error

	switch p.lexer.Peek().Type {
	case lexer.Ident:
		child, err = parseTypeIdent(p)
	case lexer.BracketL:
		child, err = parseTypeList(p)
	default:
		return nil, makeSyntaxError(p.lexer.Peek(), "unexpected symbol", true)
	}

	if err != nil {
		return nil, err
	}

	for p.lexer.Peek().Type == lexer.Question {
		child, _ = parseTypeOptional(p, child)
	}

	return child, nil
}

func parseTypeIdent(p *Parser) (TypeSig, error) {
	if p.peekTokenIsNot(lexer.Ident) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected identifier", false)
	}

	tok := p.lexer.Next()
	return TypeIdent{tok, tok.Lexeme}, nil
}

func parseTypeList(p *Parser) (TypeSig, error) {
	if p.peekTokenIsNot(lexer.BracketL) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected left bracket", false)
	}

	tok := p.lexer.Next()
	child, err := parseTypeSig(p)
	if err != nil {
		return nil, err
	}

	if p.peekTokenIsNot(lexer.BracketR) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected right bracket", false)
	}

	p.lexer.Next()
	return TypeList{tok, child}, nil
}

func parseTypeOptional(p *Parser, child TypeSig) (TypeSig, error) {
	if p.peekTokenIsNot(lexer.Question) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected question mark", false)
	}

	tok := p.lexer.Next()
	return TypeOptional{tok, child}, nil
}

func parseExpr(p *Parser, level Precedence) (Expr, error) {
	prefix, exists := p.prefixParseFuncs[p.lexer.Peek().Type]
	if exists == false {
		return nil, makeSyntaxError(p.lexer.Peek(), "unexpected symbol", true)
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

func parseFunction(p *Parser) (Expr, error) {
	if p.peekTokenIsNot(lexer.Fn) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected FN keyword", false)
	}
	tok := p.lexer.Next()

	params, ret, err := parseFunctionSignature(p)
	if err != nil {
		return nil, err
	}

	block, err := parseStmtBlock(p)
	if err != nil {
		return nil, err
	}

	return FunctionExpr{tok, params, ret, block}, nil
}

func parseFunctionSignature(p *Parser) ([]FunctionParam, TypeSig, error) {
	var params []FunctionParam
	var ret TypeSig
	var err error

	if params, err = parseFunctionParams(p); err != nil {
		return nil, nil, err
	}

	if ret, err = parseFunctionReturnSig(p); err != nil {
		return nil, nil, err
	}

	return params, ret, nil
}

func parseFunctionParams(p *Parser) ([]FunctionParam, error) {
	if p.peekTokenIsNot(lexer.ParenL) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected left paren", false)
	}
	p.lexer.Next()

	params := []FunctionParam{}
	for p.peekTokenIsNot(lexer.ParenR, lexer.EOF, lexer.Error) {
		param, err := parseFunctionParam(p)
		if err != nil {
			return nil, err
		}

		params = append(params, param)

		if p.peekTokenIsNot(lexer.Comma) {
			break
		} else {
			p.lexer.Next()
		}
	}

	if p.peekTokenIsNot(lexer.ParenR) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected right paren", false)
	}
	p.lexer.Next()

	return params, nil
}

func parseFunctionParam(p *Parser) (FunctionParam, error) {
	ident, err := parseIdent(p)
	if err != nil {
		return FunctionParam{}, err
	}

	var sig TypeSig
	if p.lexer.Peek().Type == lexer.Colon {
		p.lexer.Next()
		sig, err = parseTypeSig(p)
		if err != nil {
			return FunctionParam{}, err
		}
	}

	return FunctionParam{ident.(IdentExpr), sig}, nil
}

func parseFunctionReturnSig(p *Parser) (sig TypeSig, err error) {
	if p.lexer.Peek().Type == lexer.Colon {
		p.lexer.Next()
		sig, err = parseTypeSig(p)
		if err != nil {
			return nil, err
		}
	}

	return sig, nil
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
		return nil, makeSyntaxError(p.lexer.Peek(), "expected left paren", false)
	}

	p.lexer.Next()
	expr, err := parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	if p.peekTokenIsNot(lexer.ParenR) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected right paren", false)
	}

	p.lexer.Next()
	return expr, nil
}

func parseIdent(p *Parser) (Expr, error) {
	if p.peekTokenIsNot(lexer.Ident) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected identifier", false)
	}

	tok := p.lexer.Next()
	return IdentExpr{tok, tok.Lexeme}, nil
}

func parseNumber(p *Parser) (Expr, error) {
	if p.peekTokenIsNot(lexer.Number) {
		return nil, makeSyntaxError(p.lexer.Peek(), "expected number literal", false)
	}

	tok := p.lexer.Next()
	return evalNumber(tok)
}

func evalNumber(tok lexer.Token) (NumberExpr, error) {
	val, err := strconv.ParseUint(tok.Lexeme, 10, 64)
	if err != nil {
		return NumberExpr{}, makeSyntaxError(tok, "malformed number literal", false)
	}

	return NumberExpr{tok, int(val)}, nil
}
