package lang

import (
	"strconv"
	"strings"
)

// precedence describes the relative binding powers of different operators
type precedence int

// The staticly defined precedence levels
const (
	precLowest precedence = iota * 10
	precAssign
	precComparison
	precSum
	precProduct
	precPrefix
	precPostfix
	precDispatch
)

// prefixParseFunc describes the parsing function for any construct where the
// binding operator comes before the expression it binds to.
type prefixParseFunc func(p *parser) (Expr, error)

// postfixParseFunc describes the parsing function for any construct where the
// binding operator comes after the expression it binds to.
type postfixParseFunc func(p *parser, left Expr) (Expr, error)

// parser contains methods for generating an abstract syntax tree from a
// sequence of Tokens
type parser struct {
	lexer             *lexer
	funcDepth         int
	precedenceTable   map[tokType]precedence
	prefixParseFuncs  map[tokType]prefixParseFunc
	postfixParseFuncs map[tokType]postfixParseFunc
}

func (p *parser) errorFromPeekToken(message string) error {
	return p.errorFromLocation(p.lexer.peek().Loc, message)
}

func (p *parser) errorFromLocation(loc Loc, message string) error {
	return SyntaxError{p.lexer.Filepath, loc, message}
}

func (p *parser) peekTokenIsNot(first tokType, rest ...tokType) bool {
	peek := p.lexer.peek().Type

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

func (p *parser) expectNextToken(which tokType, otherwise string) (token, error) {
	if p.peekTokenIsNot(which) {
		peek := p.lexer.peek()
		return peek, p.errorFromPeekToken(otherwise)
	}

	return p.lexer.next(), nil
}

func (p *parser) registerPrecedence(typ tokType, level precedence) {
	p.precedenceTable[typ] = level
}

func (p *parser) registerPrefix(typ tokType, fn prefixParseFunc) {
	p.prefixParseFuncs[typ] = fn
}

func (p *parser) registerPostfix(typ tokType, fn postfixParseFunc, level precedence) {
	p.registerPrecedence(typ, level)
	p.postfixParseFuncs[typ] = fn
}

func (p *parser) peekPrecedence() precedence {
	prec, exists := p.precedenceTable[p.lexer.peek().Type]

	if exists {
		return prec
	}

	return precLowest
}

// Parse initializers a parser and defines the grammar precedence levels
func Parse(filepath string, source string) (*RootNode, error) {
	p := makeParser(filepath, source)
	loadGrammar(p)
	return parseProgram(p)
}

func makeParser(filepath string, source string) *parser {
	s := scan(source)
	l := Lex(filepath, s)
	p := &parser{
		l,
		0,
		make(map[tokType]precedence),
		make(map[tokType]prefixParseFunc),
		make(map[tokType]postfixParseFunc),
	}

	return p
}

func loadGrammar(p *parser) {
	p.registerPrefix(tokFn, parseFunction)
	p.registerPrefix(tokBracketL, parseList)
	p.registerPrefix(tokParenL, parseGroup)
	p.registerPrefix(tokPlus, parsePrefix)
	p.registerPrefix(tokDash, parsePrefix)
	p.registerPrefix(tokSelf, parseSelf)
	p.registerPrefix(tokIdent, parseIdent)
	p.registerPrefix(tokNumber, parseNumber)
	p.registerPrefix(tokString, parseString)
	p.registerPrefix(tokBoolean, parseBoolean)

	p.registerPostfix(tokBracketL, parseSubscript, precDispatch)
	p.registerPostfix(tokParenL, parseDispatch, precDispatch)
	p.registerPostfix(tokAssign, parseAssign, precAssign)
	p.registerPostfix(tokLT, parseInfix, precComparison)
	p.registerPostfix(tokLTEquals, parseInfix, precComparison)
	p.registerPostfix(tokGT, parseInfix, precComparison)
	p.registerPostfix(tokGTEquals, parseInfix, precComparison)
	p.registerPostfix(tokPlus, parseInfix, precSum)
	p.registerPostfix(tokDash, parseInfix, precSum)
	p.registerPostfix(tokStar, parseInfix, precProduct)
	p.registerPostfix(tokSlash, parseInfix, precProduct)
}

func parseProgram(p *parser) (*RootNode, error) {
	stmts := []Stmt{}

	for p.peekTokenIsNot(tokError, tokEOF) {
		var stmt Stmt
		var err error
		switch p.lexer.peek().Type {
		case tokUse:
			stmt, err = parseUseStmt(p)
		case tokPub:
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

func parseStmt(p *parser) (Stmt, error) {
	if p.funcDepth == 0 {
		return parseTopLevelStmt(p)
	}

	return parseNonTopLevelStmt(p)
}

func parseTopLevelStmt(p *parser) (Stmt, error) {
	switch p.lexer.peek().Type {
	case tokReturn:
		return nil, p.errorFromPeekToken("return statements must be inside a function")
	default:
		return parseGeneralStmt(p)
	}
}

func parseNonTopLevelStmt(p *parser) (Stmt, error) {
	switch p.lexer.peek().Type {
	case tokReturn:
		return parseReturnStmt(p)
	default:
		return parseGeneralStmt(p)
	}
}

func parseGeneralStmt(p *parser) (Stmt, error) {
	switch p.lexer.peek().Type {
	case tokUse:
		return nil, p.errorFromPeekToken("use statements must be outside any other statement")
	case tokIf:
		return parseIfStmt(p)
	case tokLet:
		return parseDeclarationStmt(p)
	default:
		return parseExprStmt(p)
	}
}

func parseStmtBlock(p *parser) (*StmtBlock, error) {
	left, err := p.expectNextToken(tokBraceL, "expected left brace")
	if err != nil {
		return &StmtBlock{}, err
	}

	stmts := []Stmt{}
	for p.peekTokenIsNot(tokBraceR, tokEOF, tokError) {
		var stmt Stmt
		stmt, err = parseStmt(p)
		if err != nil {
			return &StmtBlock{}, err
		}

		stmts = append(stmts, stmt)
	}

	right, err := p.expectNextToken(tokBraceR, "expected right brace")
	if err != nil {
		return &StmtBlock{}, err
	}

	return &StmtBlock{left, stmts, right}, nil
}

func parseUseStmt(p *parser) (Stmt, error) {
	tok, err := p.expectNextToken(tokUse, "expected USE keyword")
	if err != nil {
		return nil, err
	}

	var expr Expr
	if expr, err = parseString(p); err != nil {
		return nil, err
	}

	path := expr.(*StringExpr)

	var filter []*UseFilter
	if p.lexer.peek().Type == tokParenL {
		filter, err = parseUseFilters(p)
	}

	_, err = p.expectNextToken(tokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return &UseStmt{tok, path, filter}, nil
}

func parseUseFilters(p *parser) (filter []*UseFilter, err error) {
	_, err = p.expectNextToken(tokParenL, "expected left paren")
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
		if p.lexer.peek().Type != tokIdent {
			break
		}

		// 2.
		expr, _ := parseIdent(p)
		name := expr.(*IdentExpr)
		filter = append(filter, &UseFilter{name})

		// 3.
		if p.lexer.peek().Type == tokComma {
			p.lexer.next()
			continue
		}

		// 4.
		break
	}

	_, err = p.expectNextToken(tokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return filter, nil
}

func parsePubStmt(p *parser) (Stmt, error) {
	tok, err := p.expectNextToken(tokPub, "expected PUB keyword")
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

func parseIfStmt(p *parser) (Stmt, error) {
	tok, err := p.expectNextToken(tokIf, "expected IF keyword")
	if err != nil {
		return nil, err
	}

	var cond Expr
	if cond, err = parseExpr(p, precLowest); err != nil {
		return nil, err
	}

	var clause *StmtBlock
	if clause, err = parseStmtBlock(p); err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(tokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return &IfStmt{tok, cond, clause}, nil
}

func parseDeclarationStmt(p *parser) (Stmt, error) {
	tok, err := p.expectNextToken(tokLet, "expected LET keyword")
	if err != nil {
		return nil, err
	}

	var expr Expr
	if expr, err = parseIdent(p); err != nil {
		return nil, err
	}

	name := expr.(*IdentExpr)

	_, err = p.expectNextToken(tokAssign, "expected :=")
	if err != nil {
		return nil, err
	}

	if expr, err = parseExpr(p, precLowest); err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(tokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return &DeclarationStmt{tok, name, expr}, nil
}

func parseReturnStmt(p *parser) (Stmt, error) {
	tok, err := p.expectNextToken(tokReturn, "expected RETURN keyword")
	if err != nil {
		return nil, err
	}

	var expr Expr
	if p.peekTokenIsNot(tokSemi, tokEOF, tokError) {
		expr, err = parseExpr(p, precLowest)
		if err != nil {
			return nil, err
		}
	}

	_, err = p.expectNextToken(tokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return &ReturnStmt{tok, expr}, nil
}

func parseExprStmt(p *parser) (Stmt, error) {
	expr, err := parseExpr(p, precLowest)
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

	_, err = p.expectNextToken(tokSemi, "expected semicolon")
	if err != nil {
		return nil, err
	}

	return stmt, nil
}

func parseTypeNote(p *parser) (TypeNote, error) {
	var child TypeNote
	var err error

	switch p.lexer.peek().Type {
	case tokIdent:
		child, err = parseTypeNoteIdent(p)
	case tokBracketL:
		child, err = parseTypeNoteList(p)
	case tokParenL:
		child, err = parseTypeNoteTuple(p)
	case tokError:
		return nil, p.errorFromPeekToken(p.lexer.peek().Lexeme)
	default:
		return nil, p.errorFromPeekToken("unexpected symbol")
	}

	if err != nil {
		return nil, err
	}

	for p.lexer.peek().Type == tokQuestion {
		child, _ = parseTypeNoteOptional(p, child)
	}

	return child, nil
}

func parseTypeNoteIdent(p *parser) (TypeNote, error) {
	var tok token
	var err error
	if tok, err = p.expectNextToken(tokIdent, "expected identifier"); err != nil {
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

func parseTypeNoteList(p *parser) (TypeNote, error) {
	tok, err := p.expectNextToken(tokBracketL, "expected left bracket")
	if err != nil {
		return nil, err
	}

	child, err := parseTypeNote(p)
	if err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(tokBracketR, "expected right bracket")
	if err != nil {
		return nil, err
	}

	return TypeNoteList{tok, child}, nil
}

func parseTypeNoteOptional(p *parser, child TypeNote) (TypeNote, error) {
	tok, err := p.expectNextToken(tokQuestion, "expected question mark")
	if err != nil {
		return nil, err
	}

	return TypeNoteOptional{tok, child}, nil
}

func parseTypeNoteTuple(p *parser) (TypeNote, error) {
	tok, err := p.expectNextToken(tokParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	params := []TypeNote{}
	for p.peekTokenIsNot(tokParenR, tokError, tokEOF) {
		var sig TypeNote
		sig, err = parseTypeNote(p)
		if err != nil {
			return nil, err
		}

		params = append(params, sig)

		if p.peekTokenIsNot(tokComma) {
			break
		} else {
			p.lexer.next()
		}
	}

	_, err = p.expectNextToken(tokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	tuple := TypeNoteTuple{tok, params}
	if p.peekTokenIsNot(tokArrow) {
		return tuple, nil
	}

	return parseTypeNoteFunction(p, tuple)
}

func parseTypeNoteFunction(p *parser, tuple TypeNoteTuple) (TypeNote, error) {
	_, err := p.expectNextToken(tokArrow, "expected arrow")
	if err != nil {
		return nil, err
	}

	ret, err := parseTypeNote(p)
	if err != nil {
		return nil, err
	}

	return TypeNoteFunction{tuple, ret}, nil
}

func parseExpr(p *parser, level precedence) (Expr, error) {
	prefix, exists := p.prefixParseFuncs[p.lexer.peek().Type]
	if exists == false {
		peek := p.lexer.peek()
		if peek.Type == tokError {
			return nil, p.errorFromPeekToken(peek.Lexeme)
		}
		return nil, p.errorFromPeekToken("unexpected symbol")
	}

	left, err := prefix(p)
	if err != nil {
		return nil, err
	}

	for p.peekTokenIsNot(tokEOF) && level < p.peekPrecedence() {
		infix := p.postfixParseFuncs[p.lexer.peek().Type]
		left, err = infix(p, left)

		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func parseFunction(p *parser) (Expr, error) {
	tok, err := p.expectNextToken(tokFn, "expected FN keyword")
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

func parseFunctionSignature(p *parser) ([]*FunctionParam, TypeNote, error) {
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

func parseFunctionParams(p *parser) ([]*FunctionParam, error) {
	_, err := p.expectNextToken(tokParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	params := []*FunctionParam{}
	for p.peekTokenIsNot(tokParenR, tokEOF, tokError) {
		var param *FunctionParam
		param, err = parseFunctionParam(p)
		if err != nil {
			return nil, err
		}

		params = append(params, param)

		if p.peekTokenIsNot(tokComma) {
			break
		} else {
			p.lexer.next()
		}
	}

	_, err = p.expectNextToken(tokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return params, nil
}

func parseFunctionParam(p *parser) (*FunctionParam, error) {
	ident, err := parseIdent(p)
	if err != nil {
		return &FunctionParam{}, err
	}

	_, err = p.expectNextToken(tokColon, "expected colon between parameter name and type")
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

func parseFunctionReturnSig(p *parser) (TypeNote, error) {
	_, err := p.expectNextToken(tokColon, "expected colon between parameters and return type")
	if err != nil {
		return nil, err
	}

	ret, err := parseTypeNote(p)
	if err != nil {
		return nil, err
	}

	return ret, err
}

func parseInfix(p *parser, left Expr) (Expr, error) {
	level := p.peekPrecedence()
	tok := p.lexer.next()
	oper := tok.Lexeme
	right, err := parseExpr(p, level)
	if err != nil {
		return nil, err
	}

	return &BinaryExpr{oper, tok, left, right}, nil
}

func parseList(p *parser) (Expr, error) {
	tok, err := p.expectNextToken(tokBracketL, "expected left bracket")
	if err != nil {
		return nil, err
	}

	elements := []Expr{}
	for p.peekTokenIsNot(tokBracketR, tokError, tokEOF) {
		var elem Expr
		elem, err = parseExpr(p, precLowest)
		if err != nil {
			return nil, err
		}

		elements = append(elements, elem)

		if p.peekTokenIsNot(tokComma) {
			break
		}

		p.lexer.next()
	}

	_, err = p.expectNextToken(tokBracketR, "expected right bracket")
	if err != nil {
		return nil, err
	}

	return &ListExpr{tok, elements}, nil
}

func parseSubscript(p *parser, left Expr) (Expr, error) {
	_, err := p.expectNextToken(tokBracketL, "expect left bracket")
	if err != nil {
		return nil, err
	}

	if p.lexer.peek().Type == tokBracketR {
		return nil, p.errorFromPeekToken("expected index expression")
	}

	index, err := parseExpr(p, precLowest)
	if err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(tokBracketR, "expect right bracket")
	if err != nil {
		return nil, err
	}

	return &SubscriptExpr{left, index}, nil
}

func parseDispatch(p *parser, left Expr) (Expr, error) {
	_, err := p.expectNextToken(tokParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	var args []Expr
	for p.peekTokenIsNot(tokParenR, tokError, tokEOF) {
		var arg Expr
		arg, err = parseExpr(p, precLowest)
		if err != nil {
			return nil, err
		}

		args = append(args, arg)
		if p.peekTokenIsNot(tokComma) {
			break
		}

		p.lexer.next()
	}

	_, err = p.expectNextToken(tokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return &DispatchExpr{left, args}, nil
}

func parseAssign(p *parser, left Expr) (Expr, error) {
	leftIdent, ok := left.(*IdentExpr)
	if ok == false {
		return nil, p.errorFromLocation(left.Start(), "left hand must be an identifier")
	}

	level := p.peekPrecedence()
	tok := p.lexer.next()
	right, err := parseExpr(p, level-1)
	if err != nil {
		return nil, err
	}

	return &AssignExpr{tok, leftIdent, right}, nil
}

func parsePostfix(p *parser, left Expr) (Expr, error) {
	tok := p.lexer.next()
	oper := tok.Lexeme

	return &UnaryExpr{oper, tok, left}, nil
}

func parsePrefix(p *parser) (Expr, error) {
	tok := p.lexer.next()
	oper := tok.Lexeme
	right, err := parseExpr(p, precPrefix)
	if err != nil {
		return nil, err
	}

	return &UnaryExpr{oper, tok, right}, nil
}

func parseGroup(p *parser) (Expr, error) {
	_, err := p.expectNextToken(tokParenL, "expected left paren")
	if err != nil {
		return nil, err
	}

	expr, err := parseExpr(p, precLowest)
	if err != nil {
		return nil, err
	}

	_, err = p.expectNextToken(tokParenR, "expected right paren")
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func parseSelf(p *parser) (Expr, error) {
	tok, err := p.expectNextToken(tokSelf, "expected self")
	if err != nil {
		return nil, err
	}

	return &SelfExpr{tok}, nil
}

func parseIdent(p *parser) (Expr, error) {
	tok, err := p.expectNextToken(tokIdent, "expected identifier")
	if err != nil {
		return nil, err
	}

	return &IdentExpr{tok, tok.Lexeme}, nil
}

func parseNumber(p *parser) (Expr, error) {
	tok, err := p.expectNextToken(tokNumber, "expected number literal")
	if err != nil {
		return nil, err
	}

	return evalNumber(p, tok)
}

func evalNumber(p *parser, tok token) (*NumberExpr, error) {
	val, err := strconv.ParseUint(tok.Lexeme, 10, 64)
	if err != nil {
		return nil, p.errorFromLocation(tok.Loc, "malformed number literal")
	}

	return &NumberExpr{tok, int(val)}, nil
}

func parseString(p *parser) (Expr, error) {
	tok, err := p.expectNextToken(tokString, "expected string literal")
	if err != nil {
		return nil, err
	}

	return evalString(tok)
}

func evalString(tok token) (*StringExpr, error) {
	dblQuote := "\""
	remSuffix := strings.TrimSuffix(tok.Lexeme, dblQuote)
	remBoth := strings.TrimPrefix(remSuffix, dblQuote)

	return &StringExpr{tok, remBoth}, nil
}

func parseBoolean(p *parser) (Expr, error) {
	tok, err := p.expectNextToken(tokBoolean, "expected boolean literal")
	if err != nil {
		return nil, err
	}

	return evalBoolean(p, tok)
}

func evalBoolean(p *parser, tok token) (*BooleanExpr, error) {
	if tok.Lexeme == "true" {
		return &BooleanExpr{tok, true}, nil
	} else if tok.Lexeme == "false" {
		return &BooleanExpr{tok, false}, nil
	}

	return nil, p.errorFromLocation(tok.Loc, "malformed boolean literal")
}
