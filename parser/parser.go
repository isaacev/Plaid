package parser

import (
	"fmt"
	"plaid/lexer"
	"strconv"
	"strings"
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
	Lowest Precedence = iota * 10
	Assign
	Sum
	Product
	Prefix
	Postfix
	Dispatch
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

func (p *Parser) expectNextToken(which lexer.Type, otherwise string) (lexer.Token, error) {
	if p.peekTokenIsNot(which) {
		peek := p.lexer.Peek()
		return peek, makeSyntaxError(peek, otherwise, false)
	}

	return p.lexer.Next(), nil
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
	p.registerPrefix(lexer.String, parseString)
	p.registerPrefix(lexer.Boolean, parseBoolean)

	p.registerPostfix(lexer.BracketL, parseSubscript, Dispatch)
	p.registerPostfix(lexer.ParenL, parseDispatch, Dispatch)
	p.registerPostfix(lexer.Assign, parseAssign, Assign)
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
	case lexer.Return:
		return parseReturnStmt(p)
	default:
		return parseExprStmt(p)
	}
}

func parseStmtBlock(p *Parser) (StmtBlock, error) {
	left, err := p.expectNextToken(lexer.BraceL, "expected left brace")
	if err != nil {
		return StmtBlock{}, err
	}

	stmts := []Stmt{}
	for p.peekTokenIsNot(lexer.BraceR, lexer.EOF, lexer.Error) {
		var stmt Stmt
		stmt, err = parseStmt(p)
		if err != nil {
			return StmtBlock{}, err
		}

		stmts = append(stmts, stmt)
	}

	right, err := p.expectNextToken(lexer.BraceR, "expected right brace")
	if err != nil {
		return StmtBlock{}, err
	}

	return StmtBlock{left, stmts, right}, nil
}

func parseDeclarationStmt(p *Parser) (Stmt, error) {
	tok, err := p.expectNextToken(lexer.Let, "expected LET keyword")
	if err != nil {
		return nil, err
	}

	var expr Expr
	if expr, err = parseIdent(p); err != nil {
		return nil, err
	}

	name := expr.(IdentExpr)

	_, err = p.expectNextToken(lexer.Assign, "expected :=")
	if err != nil {
		return nil, err
	}

	if expr, err = parseExpr(p, Lowest); err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(lexer.Semi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return DeclarationStmt{tok, name, expr}, nil
}

func parseReturnStmt(p *Parser) (Stmt, error) {
	tok, err := p.expectNextToken(lexer.Return, "expected RETURN keyword")
	if err != nil {
		return nil, err
	}

	var expr Expr
	if p.peekTokenIsNot(lexer.Semi, lexer.EOF, lexer.Error) {
		expr, err = parseExpr(p, Lowest)
		if err != nil {
			return nil, err
		}
	}

	_, err = p.expectNextToken(lexer.Semi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return ReturnStmt{tok, expr}, nil
}

func parseExprStmt(p *Parser) (Stmt, error) {
	expr, err := parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	var stmt Stmt
	switch expr.(type) {
	case DispatchExpr:
		stmt = ExprStmt{expr}
	case AssignExpr:
		stmt = ExprStmt{expr}
	default:
		return nil, SyntaxError{expr.Start(), "expected start of statement"}
	}

	_, err = p.expectNextToken(lexer.Semi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func parseTypeNote(p *Parser) (TypeNote, error) {
	var child TypeNote
	var err error

	switch p.lexer.Peek().Type {
	case lexer.Ident:
		child, err = parseTypeNoteIdent(p)
	case lexer.BracketL:
		child, err = parseTypeNoteList(p)
	case lexer.ParenL:
		child, err = parseTypeNoteTuple(p)
	default:
		return nil, makeSyntaxError(p.lexer.Peek(), "unexpected symbol", true)
	}

	if err != nil {
		return nil, err
	}

	for p.lexer.Peek().Type == lexer.Question {
		child, _ = parseTypeNoteOptional(p, child)
	}

	return child, nil
}

func parseTypeNoteIdent(p *Parser) (TypeNote, error) {
	var tok lexer.Token
	var err error
	if tok, err = p.expectNextToken(lexer.Ident, "expected identifier"); err != nil {
		return nil, err
	}

	switch tok.Lexeme {
	case "Void":
		return TypeNoteVoid{tok}, nil
	default:
		return TypeNoteIdent{tok, tok.Lexeme}, nil
	}
}

func parseTypeNoteList(p *Parser) (TypeNote, error) {
	tok, err := p.expectNextToken(lexer.BracketL, "expected left bracket")
	if err != nil {
		return nil, err
	}

	child, err := parseTypeNote(p)
	if err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(lexer.BracketR, "expected right bracket")
	if err != nil {
		return nil, err
	}

	return TypeNoteList{tok, child}, nil
}

func parseTypeNoteOptional(p *Parser, child TypeNote) (TypeNote, error) {
	tok, err := p.expectNextToken(lexer.Question, "expected question mark")
	if err != nil {
		return nil, err
	}

	return TypeNoteOptional{tok, child}, nil
}

func parseTypeNoteTuple(p *Parser) (TypeNote, error) {
	tok, err := p.expectNextToken(lexer.ParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	params := []TypeNote{}
	for p.peekTokenIsNot(lexer.ParenR, lexer.Error, lexer.EOF) {
		var sig TypeNote
		sig, err = parseTypeNote(p)
		if err != nil {
			return nil, err
		}

		params = append(params, sig)

		if p.peekTokenIsNot(lexer.Comma) {
			break
		} else {
			p.lexer.Next()
		}
	}

	_, err = p.expectNextToken(lexer.ParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	tuple := TypeNoteTuple{tok, params}
	if p.peekTokenIsNot(lexer.Arrow) {
		return tuple, nil
	}

	return parseTypeNoteFunction(p, tuple)
}

func parseTypeNoteFunction(p *Parser, tuple TypeNoteTuple) (TypeNote, error) {
	_, err := p.expectNextToken(lexer.Arrow, "expected arrow")
	if err != nil {
		return nil, err
	}

	ret, err := parseTypeNote(p)
	if err != nil {
		return nil, err
	}

	return TypeNoteFunction{tuple, ret}, nil
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
	tok, err := p.expectNextToken(lexer.Fn, "expected FN keyword")
	if err != nil {
		return nil, err
	}

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

func parseFunctionSignature(p *Parser) ([]FunctionParam, TypeNote, error) {
	var params []FunctionParam
	var ret TypeNote
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
	_, err := p.expectNextToken(lexer.ParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	params := []FunctionParam{}
	for p.peekTokenIsNot(lexer.ParenR, lexer.EOF, lexer.Error) {
		var param FunctionParam
		param, err = parseFunctionParam(p)
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

	_, err = p.expectNextToken(lexer.ParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return params, nil
}

func parseFunctionParam(p *Parser) (FunctionParam, error) {
	ident, err := parseIdent(p)
	if err != nil {
		return FunctionParam{}, err
	}

	_, err = p.expectNextToken(lexer.Colon, "expected colon between parameter name and type")
	if err != nil {
		return FunctionParam{}, err
	}

	var sig TypeNote
	sig, err = parseTypeNote(p)
	if err != nil {
		return FunctionParam{}, err
	}

	return FunctionParam{ident.(IdentExpr), sig}, nil
}

func parseFunctionReturnSig(p *Parser) (TypeNote, error) {
	_, err := p.expectNextToken(lexer.Colon, "expected colon between parameters and return type")
	if err != nil {
		return nil, err
	}

	ret, err := parseTypeNote(p)
	if err != nil {
		return nil, err
	}

	return ret, err
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

func parseSubscript(p *Parser, left Expr) (Expr, error) {
	_, err := p.expectNextToken(lexer.BracketL, "expect left bracket")
	if err != nil {
		return nil, err
	}

	index, err := parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(lexer.BracketR, "expect right bracket")
	if err != nil {
		return nil, err
	}

	return SubscriptExpr{left, index}, nil
}

func parseDispatch(p *Parser, left Expr) (Expr, error) {
	_, err := p.expectNextToken(lexer.ParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	var args []Expr
	for p.peekTokenIsNot(lexer.ParenR) {
		var arg Expr
		arg, err = parseExpr(p, Lowest)
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
		if p.peekTokenIsNot(lexer.Comma) {
			break
		}

		p.lexer.Next()
	}

	_, err = p.expectNextToken(lexer.ParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return DispatchExpr{left, args}, nil
}

func parseAssign(p *Parser, left Expr) (Expr, error) {
	leftIdent, ok := left.(IdentExpr)
	if ok == false {
		return nil, SyntaxError{left.Start(), "left hand must be an identifier"}
	}

	level := p.peekPrecedence()
	tok := p.lexer.Next()
	right, err := parseExpr(p, level-1)
	if err != nil {
		return nil, err
	}

	return AssignExpr{tok, leftIdent, right}, nil
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
	_, err := p.expectNextToken(lexer.ParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	expr, err := parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(lexer.ParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func parseIdent(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(lexer.Ident, "expected identifier")
	if err != nil {
		return nil, err
	}

	return IdentExpr{tok, tok.Lexeme}, nil
}

func parseNumber(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(lexer.Number, "expected number literal")
	if err != nil {
		return nil, err
	}

	return evalNumber(tok)
}

func evalNumber(tok lexer.Token) (NumberExpr, error) {
	val, err := strconv.ParseUint(tok.Lexeme, 10, 64)
	if err != nil {
		return NumberExpr{}, makeSyntaxError(tok, "malformed number literal", false)
	}

	return NumberExpr{tok, int(val)}, nil
}

func parseString(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(lexer.String, "expected string literal")
	if err != nil {
		return nil, err
	}

	return evalString(tok)
}

func evalString(tok lexer.Token) (StringExpr, error) {
	dblQuote := "\""
	remSuffix := strings.TrimSuffix(tok.Lexeme, dblQuote)
	remBoth := strings.TrimPrefix(remSuffix, dblQuote)

	return StringExpr{tok, remBoth}, nil
}

func parseBoolean(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(lexer.Boolean, "expected boolean literal")
	if err != nil {
		return nil, err
	}

	return evalBoolean(tok)
}

func evalBoolean(tok lexer.Token) (BooleanExpr, error) {
	if tok.Lexeme == "true" {
		return BooleanExpr{tok, true}, nil
	} else if tok.Lexeme == "false" {
		return BooleanExpr{tok, false}, nil
	}

	return BooleanExpr{}, makeSyntaxError(tok, "malformed boolean literal", false)
}
