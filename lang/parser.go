package lang

import (
	"strconv"
	"strings"
)

// Precedence describes the relative binding powers of different operators
type Precedence int

// The staticly defined precedence levels
const (
	Lowest Precedence = iota * 10
	Assign
	Comparison
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
	lexer             *Lexer
	funcDepth         int
	precedenceTable   map[TokenType]Precedence
	prefixParseFuncs  map[TokenType]PrefixParseFunc
	postfixParseFuncs map[TokenType]PostfixParseFunc
}

func (p *Parser) errorFromPeekToken(message string) error {
	return p.errorFromLocation(p.lexer.Peek().Loc, message)
}

func (p *Parser) errorFromLocation(loc Loc, message string) error {
	return SyntaxError{p.lexer.Filepath, loc, message}
}

func (p *Parser) peekTokenIsNot(first TokenType, rest ...TokenType) bool {
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

func (p *Parser) expectNextToken(which TokenType, otherwise string) (Token, error) {
	if p.peekTokenIsNot(which) {
		peek := p.lexer.Peek()
		return peek, p.errorFromPeekToken(otherwise)
	}

	return p.lexer.Next(), nil
}

func (p *Parser) registerPrecedence(typ TokenType, level Precedence) {
	p.precedenceTable[typ] = level
}

func (p *Parser) registerPrefix(typ TokenType, fn PrefixParseFunc) {
	p.prefixParseFuncs[typ] = fn
}

func (p *Parser) registerPostfix(typ TokenType, fn PostfixParseFunc, level Precedence) {
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
func Parse(filepath string, source string) (*RootNode, error) {
	p := makeParser(filepath, source)
	loadGrammar(p)
	return parseProgram(p)
}

func makeParser(filepath string, source string) *Parser {
	s := Scan(source)
	l := Lex(filepath, s)
	p := &Parser{
		l,
		0,
		make(map[TokenType]Precedence),
		make(map[TokenType]PrefixParseFunc),
		make(map[TokenType]PostfixParseFunc),
	}

	return p
}

func loadGrammar(p *Parser) {
	p.registerPrefix(TokFn, parseFunction)
	p.registerPrefix(TokBracketL, parseList)
	p.registerPrefix(TokParenL, parseGroup)
	p.registerPrefix(TokPlus, parsePrefix)
	p.registerPrefix(TokDash, parsePrefix)
	p.registerPrefix(TokSelf, parseSelf)
	p.registerPrefix(TokIdent, parseIdent)
	p.registerPrefix(TokNumber, parseNumber)
	p.registerPrefix(TokString, parseString)
	p.registerPrefix(TokBoolean, parseBoolean)

	p.registerPostfix(TokBracketL, parseSubscript, Dispatch)
	p.registerPostfix(TokParenL, parseDispatch, Dispatch)
	p.registerPostfix(TokAssign, parseAssign, Assign)
	p.registerPostfix(TokLT, parseInfix, Comparison)
	p.registerPostfix(TokLTEquals, parseInfix, Comparison)
	p.registerPostfix(TokGT, parseInfix, Comparison)
	p.registerPostfix(TokGTEquals, parseInfix, Comparison)
	p.registerPostfix(TokPlus, parseInfix, Sum)
	p.registerPostfix(TokDash, parseInfix, Sum)
	p.registerPostfix(TokStar, parseInfix, Product)
	p.registerPostfix(TokSlash, parseInfix, Product)
}

func parseProgram(p *Parser) (*RootNode, error) {
	stmts := []Stmt{}

	for p.peekTokenIsNot(TokError, TokEOF) {
		var stmt Stmt
		var err error
		switch p.lexer.Peek().Type {
		case TokUse:
			stmt, err = parseUseStmt(p)
		case TokPub:
			stmt, err = parsePubStmt(p)
		default:
			stmt, err = parseTopLevelStmt(p)
		}

		if err != nil {
			return &RootNode{}, err
		}

		stmts = append(stmts, stmt)
	}

	return &RootNode{stmts}, nil
}

func parseStmt(p *Parser) (Stmt, error) {
	if p.funcDepth == 0 {
		return parseTopLevelStmt(p)
	}

	return parseNonTopLevelStmt(p)
}

func parseTopLevelStmt(p *Parser) (Stmt, error) {
	switch p.lexer.Peek().Type {
	case TokReturn:
		return nil, p.errorFromPeekToken("return statements must be inside a function")
	default:
		return parseGeneralStmt(p)
	}
}

func parseNonTopLevelStmt(p *Parser) (Stmt, error) {
	switch p.lexer.Peek().Type {
	case TokReturn:
		return parseReturnStmt(p)
	default:
		return parseGeneralStmt(p)
	}
}

func parseGeneralStmt(p *Parser) (Stmt, error) {
	switch p.lexer.Peek().Type {
	case TokUse:
		return nil, p.errorFromPeekToken("use statements must be outside any other statement")
	case TokIf:
		return parseIfStmt(p)
	case TokLet:
		return parseDeclarationStmt(p)
	default:
		return parseExprStmt(p)
	}
}

func parseStmtBlock(p *Parser) (*StmtBlock, error) {
	left, err := p.expectNextToken(TokBraceL, "expected left brace")
	if err != nil {
		return &StmtBlock{}, err
	}

	stmts := []Stmt{}
	for p.peekTokenIsNot(TokBraceR, TokEOF, TokError) {
		var stmt Stmt
		stmt, err = parseStmt(p)
		if err != nil {
			return &StmtBlock{}, err
		}

		stmts = append(stmts, stmt)
	}

	right, err := p.expectNextToken(TokBraceR, "expected right brace")
	if err != nil {
		return &StmtBlock{}, err
	}

	return &StmtBlock{left, stmts, right}, nil
}

func parseUseStmt(p *Parser) (Stmt, error) {
	tok, err := p.expectNextToken(TokUse, "expected USE keyword")
	if err != nil {
		return nil, err
	}

	var expr Expr
	if expr, err = parseString(p); err != nil {
		return nil, err
	}

	path := expr.(*StringExpr)

	var filter []*UseFilter
	if p.lexer.Peek().Type == TokParenL {
		filter, err = parseUseFilters(p)
	}

	_, err = p.expectNextToken(TokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return &UseStmt{tok, path, filter}, nil
}

func parseUseFilters(p *Parser) (filter []*UseFilter, err error) {
	_, err = p.expectNextToken(TokParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	/**
	 * Filter parsing loop:
	 * 1. if token is NOT an identifier, exit loop
	 * 2. parse a single filter identifier
	 * 3. if token is comma, eat comma and goto (1)
	 * 4. exit loop
	 */
	for {
		// 1.
		if p.lexer.Peek().Type != TokIdent {
			break
		}

		// 2.
		expr, _ := parseIdent(p)
		name := expr.(*IdentExpr)
		filter = append(filter, &UseFilter{name})

		// 3.
		if p.lexer.Peek().Type == TokComma {
			p.lexer.Next()
			continue
		}

		// 4.
		break
	}

	_, err = p.expectNextToken(TokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return filter, nil
}

func parsePubStmt(p *Parser) (Stmt, error) {
	tok, err := p.expectNextToken(TokPub, "expected PUB keyword")
	if err != nil {
		return nil, err
	}

	var decl *DeclarationStmt
	if stmt, err := parseDeclarationStmt(p); err == nil {
		decl = stmt.(*DeclarationStmt)
	} else {
		return nil, err
	}

	return &PubStmt{tok, decl}, nil
}

func parseIfStmt(p *Parser) (Stmt, error) {
	tok, err := p.expectNextToken(TokIf, "expected IF keyword")
	if err != nil {
		return nil, err
	}

	var cond Expr
	if cond, err = parseExpr(p, Lowest); err != nil {
		return nil, err
	}

	var clause *StmtBlock
	if clause, err = parseStmtBlock(p); err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(TokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return &IfStmt{tok, cond, clause}, nil
}

func parseDeclarationStmt(p *Parser) (Stmt, error) {
	tok, err := p.expectNextToken(TokLet, "expected LET keyword")
	if err != nil {
		return nil, err
	}

	var expr Expr
	if expr, err = parseIdent(p); err != nil {
		return nil, err
	}

	name := expr.(*IdentExpr)

	_, err = p.expectNextToken(TokAssign, "expected :=")
	if err != nil {
		return nil, err
	}

	if expr, err = parseExpr(p, Lowest); err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(TokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return &DeclarationStmt{tok, name, expr}, nil
}

func parseReturnStmt(p *Parser) (Stmt, error) {
	tok, err := p.expectNextToken(TokReturn, "expected RETURN keyword")
	if err != nil {
		return nil, err
	}

	var expr Expr
	if p.peekTokenIsNot(TokSemi, TokEOF, TokError) {
		expr, err = parseExpr(p, Lowest)
		if err != nil {
			return nil, err
		}
	}

	_, err = p.expectNextToken(TokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return &ReturnStmt{tok, expr}, nil
}

func parseExprStmt(p *Parser) (Stmt, error) {
	expr, err := parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	var stmt Stmt
	switch expr.(type) {
	case *DispatchExpr:
		stmt = &ExprStmt{expr}
	case *AssignExpr:
		stmt = &ExprStmt{expr}
	default:
		return nil, p.errorFromLocation(expr.Start(), "expected start of statement")
	}

	_, err = p.expectNextToken(TokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func parseTypeNote(p *Parser) (TypeNote, error) {
	var child TypeNote
	var err error

	switch p.lexer.Peek().Type {
	case TokIdent:
		child, err = parseTypeNoteIdent(p)
	case TokBracketL:
		child, err = parseTypeNoteList(p)
	case TokParenL:
		child, err = parseTypeNoteTuple(p)
	case TokError:
		return nil, p.errorFromPeekToken(p.lexer.Peek().Lexeme)
	default:
		return nil, p.errorFromPeekToken("unexpected symbol")
	}

	if err != nil {
		return nil, err
	}

	for p.lexer.Peek().Type == TokQuestion {
		child, _ = parseTypeNoteOptional(p, child)
	}

	return child, nil
}

func parseTypeNoteIdent(p *Parser) (TypeNote, error) {
	var tok Token
	var err error
	if tok, err = p.expectNextToken(TokIdent, "expected identifier"); err != nil {
		return nil, err
	}

	switch tok.Lexeme {
	case "Any":
		return TypeNoteAny{tok}, nil
	case "Void":
		return TypeNoteVoid{tok}, nil
	default:
		return TypeNoteIdent{tok, tok.Lexeme}, nil
	}
}

func parseTypeNoteList(p *Parser) (TypeNote, error) {
	tok, err := p.expectNextToken(TokBracketL, "expected left bracket")
	if err != nil {
		return nil, err
	}

	child, err := parseTypeNote(p)
	if err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(TokBracketR, "expected right bracket")
	if err != nil {
		return nil, err
	}

	return TypeNoteList{tok, child}, nil
}

func parseTypeNoteOptional(p *Parser, child TypeNote) (TypeNote, error) {
	tok, err := p.expectNextToken(TokQuestion, "expected question mark")
	if err != nil {
		return nil, err
	}

	return TypeNoteOptional{tok, child}, nil
}

func parseTypeNoteTuple(p *Parser) (TypeNote, error) {
	tok, err := p.expectNextToken(TokParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	params := []TypeNote{}
	for p.peekTokenIsNot(TokParenR, TokError, TokEOF) {
		var sig TypeNote
		sig, err = parseTypeNote(p)
		if err != nil {
			return nil, err
		}

		params = append(params, sig)

		if p.peekTokenIsNot(TokComma) {
			break
		} else {
			p.lexer.Next()
		}
	}

	_, err = p.expectNextToken(TokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	tuple := TypeNoteTuple{tok, params}
	if p.peekTokenIsNot(TokArrow) {
		return tuple, nil
	}

	return parseTypeNoteFunction(p, tuple)
}

func parseTypeNoteFunction(p *Parser, tuple TypeNoteTuple) (TypeNote, error) {
	_, err := p.expectNextToken(TokArrow, "expected arrow")
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
		peek := p.lexer.Peek()
		if peek.Type == TokError {
			return nil, p.errorFromPeekToken(peek.Lexeme)
		}
		return nil, p.errorFromPeekToken("unexpected symbol")
	}

	left, err := prefix(p)
	if err != nil {
		return nil, err
	}

	for p.peekTokenIsNot(TokEOF) && level < p.peekPrecedence() {
		infix := p.postfixParseFuncs[p.lexer.Peek().Type]
		left, err = infix(p, left)

		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func parseFunction(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(TokFn, "expected FN keyword")
	if err != nil {
		return nil, err
	}

	params, ret, err := parseFunctionSignature(p)
	if err != nil {
		return nil, err
	}

	p.funcDepth++
	block, err := parseStmtBlock(p)
	if err != nil {
		return nil, err
	}
	p.funcDepth--

	return &FunctionExpr{tok, params, ret, block}, nil
}

func parseFunctionSignature(p *Parser) ([]*FunctionParam, TypeNote, error) {
	var params []*FunctionParam
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

func parseFunctionParams(p *Parser) ([]*FunctionParam, error) {
	_, err := p.expectNextToken(TokParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	params := []*FunctionParam{}
	for p.peekTokenIsNot(TokParenR, TokEOF, TokError) {
		var param *FunctionParam
		param, err = parseFunctionParam(p)
		if err != nil {
			return nil, err
		}

		params = append(params, param)

		if p.peekTokenIsNot(TokComma) {
			break
		} else {
			p.lexer.Next()
		}
	}

	_, err = p.expectNextToken(TokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return params, nil
}

func parseFunctionParam(p *Parser) (*FunctionParam, error) {
	ident, err := parseIdent(p)
	if err != nil {
		return &FunctionParam{}, err
	}

	_, err = p.expectNextToken(TokColon, "expected colon between parameter name and type")
	if err != nil {
		return &FunctionParam{}, err
	}

	var sig TypeNote
	sig, err = parseTypeNote(p)
	if err != nil {
		return &FunctionParam{}, err
	}

	return &FunctionParam{ident.(*IdentExpr), sig}, nil
}

func parseFunctionReturnSig(p *Parser) (TypeNote, error) {
	_, err := p.expectNextToken(TokColon, "expected colon between parameters and return type")
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

	return &BinaryExpr{oper, tok, left, right}, nil
}

func parseList(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(TokBracketL, "expected left bracket")
	if err != nil {
		return nil, err
	}

	elements := []Expr{}
	for p.peekTokenIsNot(TokBracketR, TokError, TokEOF) {
		var elem Expr
		elem, err = parseExpr(p, Lowest)
		if err != nil {
			return nil, err
		}

		elements = append(elements, elem)

		if p.peekTokenIsNot(TokComma) {
			break
		}

		p.lexer.Next()
	}

	_, err = p.expectNextToken(TokBracketR, "expected right bracket")
	if err != nil {
		return nil, err
	}

	return &ListExpr{tok, elements}, nil
}

func parseSubscript(p *Parser, left Expr) (Expr, error) {
	_, err := p.expectNextToken(TokBracketL, "expect left bracket")
	if err != nil {
		return nil, err
	}

	if p.lexer.Peek().Type == TokBracketR {
		return nil, p.errorFromPeekToken("expected index expression")
	}

	index, err := parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(TokBracketR, "expect right bracket")
	if err != nil {
		return nil, err
	}

	return &SubscriptExpr{left, index}, nil
}

func parseDispatch(p *Parser, left Expr) (Expr, error) {
	_, err := p.expectNextToken(TokParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	var args []Expr
	for p.peekTokenIsNot(TokParenR, TokError, TokEOF) {
		var arg Expr
		arg, err = parseExpr(p, Lowest)
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
		if p.peekTokenIsNot(TokComma) {
			break
		}

		p.lexer.Next()
	}

	_, err = p.expectNextToken(TokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return &DispatchExpr{left, args}, nil
}

func parseAssign(p *Parser, left Expr) (Expr, error) {
	leftIdent, ok := left.(*IdentExpr)
	if ok == false {
		return nil, p.errorFromLocation(left.Start(), "left hand must be an identifier")
	}

	level := p.peekPrecedence()
	tok := p.lexer.Next()
	right, err := parseExpr(p, level-1)
	if err != nil {
		return nil, err
	}

	return &AssignExpr{tok, leftIdent, right}, nil
}

func parsePostfix(p *Parser, left Expr) (Expr, error) {
	tok := p.lexer.Next()
	oper := tok.Lexeme

	return &UnaryExpr{oper, tok, left}, nil
}

func parsePrefix(p *Parser) (Expr, error) {
	tok := p.lexer.Next()
	oper := tok.Lexeme
	right, err := parseExpr(p, Prefix)
	if err != nil {
		return nil, err
	}

	return &UnaryExpr{oper, tok, right}, nil
}

func parseGroup(p *Parser) (Expr, error) {
	_, err := p.expectNextToken(TokParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	expr, err := parseExpr(p, Lowest)
	if err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(TokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func parseSelf(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(TokSelf, "expected self")
	if err != nil {
		return nil, err
	}

	return &SelfExpr{tok}, nil
}

func parseIdent(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(TokIdent, "expected identifier")
	if err != nil {
		return nil, err
	}

	return &IdentExpr{tok, tok.Lexeme}, nil
}

func parseNumber(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(TokNumber, "expected number literal")
	if err != nil {
		return nil, err
	}

	return evalNumber(p, tok)
}

func evalNumber(p *Parser, tok Token) (*NumberExpr, error) {
	val, err := strconv.ParseUint(tok.Lexeme, 10, 64)
	if err != nil {
		return nil, p.errorFromLocation(tok.Loc, "malformed number literal")
	}

	return &NumberExpr{tok, int(val)}, nil
}

func parseString(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(TokString, "expected string literal")
	if err != nil {
		return nil, err
	}

	return evalString(tok)
}

func evalString(tok Token) (*StringExpr, error) {
	dblQuote := "\""
	remSuffix := strings.TrimSuffix(tok.Lexeme, dblQuote)
	remBoth := strings.TrimPrefix(remSuffix, dblQuote)

	return &StringExpr{tok, remBoth}, nil
}

func parseBoolean(p *Parser) (Expr, error) {
	tok, err := p.expectNextToken(TokBoolean, "expected boolean literal")
	if err != nil {
		return nil, err
	}

	return evalBoolean(p, tok)
}

func evalBoolean(p *Parser, tok Token) (*BooleanExpr, error) {
	if tok.Lexeme == "true" {
		return &BooleanExpr{tok, true}, nil
	} else if tok.Lexeme == "false" {
		return &BooleanExpr{tok, false}, nil
	}

	return nil, p.errorFromLocation(tok.Loc, "malformed boolean literal")
}
