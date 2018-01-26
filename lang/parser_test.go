package lang

import (
	"testing"
)

func TestPeekTokenIsNot(t *testing.T) {
	expectTokenToMatch := func(source string, tests ...tokType) {
		p := makeParser("", source)
		got := p.peekTokenIsNot(tests[0], tests[1:]...)

		if got != false {
			t.Errorf("Expected %t, got %t\n", false, got)
		}
	}

	expectTokenToNotMatch := func(source string, tests ...tokType) {
		p := makeParser("", source)
		got := p.peekTokenIsNot(tests[0], tests[1:]...)

		if got != true {
			t.Errorf("Expected %t, got %t\n", false, got)
		}
	}

	expectTokenToMatch("abc", tokIdent)
	expectTokenToMatch("abc", tokNumber, tokIdent)
	expectTokenToMatch("123", tokNumber, tokIdent)
	expectTokenToMatch("", tokEOF, tokError)
	expectTokenToMatch("#", tokEOF, tokError)
	expectTokenToNotMatch("abc", tokEOF, tokError)
}

func TestExpectNextToken(t *testing.T) {
	p := makeParser("", "foo 123")
	tok, err := p.expectNextToken(tokIdent, "expected an identifier")
	if err != nil {
		t.Errorf("Expected no error, got %s\n", err)
	} else if tok.Type != tokIdent {
		t.Errorf("Expected '%s', got '%s'\n", tokIdent, tok.Type)
	}

	p = makeParser("", "123 foo")
	tok, err = p.expectNextToken(tokIdent, "expected an identifier")
	exp := "(1:1) expected an identifier"
	if err != nil {
		if err.Error() != exp {
			t.Errorf("Expected '%s', got '%s'\n", exp, err)
		}
	} else if tok.Type != tokIdent {
		t.Errorf("Expected an error, got '%s'\n", tok.Type)
	}
}

func TestRegisterPrecedence(t *testing.T) {
	parser := makeParser("", "")
	parser.registerPrecedence(tokIdent, precSum)

	got := parser.precedenceTable[tokIdent]
	if got != precSum {
		t.Errorf("Expected %v, got %v\n", precSum, got)
	}
}

func TestRegisterPrefix(t *testing.T) {
	parser := makeParser("", "")
	parser.registerPrefix(tokIdent, parseIdent)

	if _, exists := parser.prefixParseFuncs[tokIdent]; exists == false {
		t.Error("Expected prefix parse function, got nothing")
	}
}

func TestRegisterPostfix(t *testing.T) {
	parser := makeParser("", "")
	parser.registerPostfix(tokPlus, parseInfix, precSum)

	if _, exists := parser.postfixParseFuncs[tokPlus]; exists == false {
		t.Error("Expected postfix parse function, got nothing")
	}

	level, exists := parser.precedenceTable[tokPlus]
	if (exists == false) || (level != precSum) {
		t.Errorf("Expected Plus precedence to be %v, got %v\n", precSum, level)
	}
}

func TestPeekPrecedence(t *testing.T) {
	parser := makeParser("", "+*")
	parser.registerPrecedence(tokPlus, precSum)

	level := parser.peekPrecedence()
	if level != precSum {
		t.Errorf("Expected Plus precedence to be %v, got %v\n", precSum, level)
	}

	parser.lexer.next()

	level = parser.peekPrecedence()
	if level != precLowest {
		t.Errorf("Expected Star precedence to be %v, got %v\n", precLowest, level)
	}
}

func TestParse(t *testing.T) {
	prog, errs := Parse("", "let a := 123; let b := 456;")

	var err error = nil
	if len(errs) > 0 {
		err = errs[0]
	}

	expectNoParserErrors(t, "(let a 123)\n(let b 456)", prog, err)
	expectStart(t, prog, 1, 1)
}

func TestParseProgram(t *testing.T) {
	p := makeParser("", "let a := 123; let b := 456;")
	loadGrammar(p)
	prog, err := parseProgram(p)
	expectNoParserErrors(t, "(let a 123)\n(let b 456)", prog, err)
	expectStart(t, prog, 1, 1)

	p = makeParser("", "")
	loadGrammar(p)
	prog, err = parseProgram(p)
	expectNoParserErrors(t, "", prog, err)
	expectStart(t, prog, 1, 1)

	p = makeParser("", `use "foo"; use "bar";`)
	loadGrammar(p)
	prog, err = parseProgram(p)
	expectNoParserErrors(t, "(use \"foo\")\n(use \"bar\")", prog, err)
	expectStart(t, prog, 1, 1)

	p = makeParser("", `pub let a := 123;`)
	loadGrammar(p)
	prog, err = parseProgram(p)
	expectNoParserErrors(t, "(pub (let a 123))", prog, err)
	expectStart(t, prog, 1, 1)

	p = makeParser("", "let a = 123; let b := 456;")
	loadGrammar(p)
	prog, err = parseProgram(p)
	expectParserError(t, "(1:7) expected :=", prog, err)
}

func TestParseStmt(t *testing.T) {
	expectStmt := func(source string, ast string, fn func(*parser) (Stmt, error)) {
		parser := makeParser("", source)
		loadGrammar(parser)
		stmt, err := fn(parser)
		expectNoParserErrors(t, ast, stmt, err)
		expectStart(t, stmt, 1, 1)
	}

	expectStmtError := func(source string, msg string, fn func(*parser) (Stmt, error)) {
		parser := makeParser("", source)
		loadGrammar(parser)
		stmt, err := fn(parser)
		expectParserError(t, msg, stmt, err)
	}

	expectStmt("if a { let a := 456; };", "(if a {\n  (let a 456)})", parseGeneralStmt)
	expectStmt("let a := 123;", "(let a 123)", parseGeneralStmt)
	expectStmt("return 123;", "(return 123)", parseNonTopLevelStmt)
	expectStmtError("123 + 456", "(1:1) expected start of statement", parseStmt)
	expectStmtError("123 + 456", "(1:1) expected start of statement", parseTopLevelStmt)
	expectStmtError("123 + 456", "(1:1) expected start of statement", parseNonTopLevelStmt)
	expectStmtError("return 123;", "(1:1) return statements must be inside a function", parseTopLevelStmt)
	expectStmtError(`use 'foo';`, "(1:1) use statements must be outside any other statement", parseStmt)
}

func TestParseStmtBlock(t *testing.T) {
	expectStmtBlock := func(source string, ast string) {
		parser := makeParser("", source)
		loadGrammar(parser)
		block, err := parseStmtBlock(parser)
		expectNoParserErrors(t, ast, block, err)
		expectStart(t, block, 1, 1)
	}

	expectStmtBlockError := func(source string, msg string) {
		parser := makeParser("", source)
		loadGrammar(parser)
		block, err := parseStmtBlock(parser)
		expectParserError(t, msg, block, err)
	}

	expectStmtBlock("{ let a := 123; }", "{\n  (let a 123)}")
	expectStmtBlockError("let a := 123; }", "(1:1) expected left brace")
	expectStmtBlockError("{ let a := 123 }", "(1:16) expected semicolon")
	expectStmtBlockError("{ let a := 123;", "(1:15) expected right brace")
}

func TestParseUseStmt(t *testing.T) {
	good := func(source string, ast string) {
		p := makeParser("", source)
		loadGrammar(p)
		stmt, err := parseUseStmt(p)
		expectNoParserErrors(t, ast, stmt, err)
		expectStart(t, stmt, 1, 1)
	}

	bad := func(source string, msg string) {
		p := makeParser("", source)
		loadGrammar(p)
		stmt, err := parseUseStmt(p)
		expectParserError(t, msg, stmt, err)
	}

	good(`use "foo";`, `(use "foo")`)
	good(`use "bar.plaid";`, `(use "bar.plaid")`)
	good(`use "foo" (a);`, `(use "foo" (a))`)
	good(`use "foo" (a, b);`, `(use "foo" (a b))`)
	good(`use "foo" (a, b,);`, `(use "foo" (a b))`)

	bad(`ues "foo";`, "(1:1) expected USE keyword")
	bad(`use 123;`, "(1:5) expected string literal")
	bad(`use "foo"`, "(1:9) expected semicolon")
}

func TestParseUseFilter(t *testing.T) {
	bad := func(source string, msg string) {
		p := makeParser("", source)
		loadGrammar(p)
		_, err := parseUseFilters(p)
		expectParserError(t, msg, nil, err)
	}

	bad(`a)`, `(1:1) expected left paren`)
	bad(`(`, `(1:1) expected right paren`)
	bad(`(123`, `(1:2) expected right paren`)
	bad(`(a b)`, `(1:4) expected right paren`)
}

func TestParsePubStmt(t *testing.T) {
	good := func(source string, ast string) {
		p := makeParser("", source)
		loadGrammar(p)
		stmt, err := parsePubStmt(p)
		expectNoParserErrors(t, ast, stmt, err)
		expectStart(t, stmt, 1, 1)
	}

	bad := func(source string, msg string) {
		p := makeParser("", source)
		loadGrammar(p)
		stmt, err := parsePubStmt(p)
		expectParserError(t, msg, stmt, err)
	}

	good(`pub let a := 123;`, `(pub (let a 123))`)
	good(`pub let x := "abc";`, `(pub (let x "abc"))`)

	bad(`pbu let a := 123;`, "(1:1) expected PUB keyword")
	bad(`pub a := 123;`, "(1:5) expected LET keyword")
}

func TestParseIfStmt(t *testing.T) {
	expectIf := func(source string, ast string) {
		p := makeParser("", source)
		loadGrammar(p)
		stmt, err := parseIfStmt(p)
		expectNoParserErrors(t, ast, stmt, err)
		expectStart(t, stmt, 1, 1)
	}

	expectIfError := func(source string, msg string) {
		p := makeParser("", source)
		loadGrammar(p)
		stmt, err := parseIfStmt(p)
		expectParserError(t, msg, stmt, err)
	}

	expectIf("if true {};", "(if true {})")
	expectIf("if true { let a := 123; };", "(if true {\n  (let a 123)})")
	expectIfError("iff true { let a := 123; };", "(1:1) expected IF keyword")
	expectIfError("if let { let a := 123; };", "(1:4) unexpected symbol")
	expectIfError("if true { let a := 123 };", "(1:24) expected semicolon")
	expectIfError("if true { let a := 123; }", "(1:25) expected semicolon")
}

func TestParseDeclarationStmt(t *testing.T) {
	p := makeParser("", "let a := 123;")
	p.registerPrefix(tokNumber, parseNumber)
	stmt, err := parseDeclarationStmt(p)
	expectNoParserErrors(t, "(let a 123)", stmt, err)
	expectStart(t, stmt, 1, 1)

	p = makeParser("", "a := 123;")
	p.registerPrefix(tokNumber, parseNumber)
	stmt, err = parseDeclarationStmt(p)
	expectParserError(t, "(1:1) expected LET keyword", stmt, err)

	p = makeParser("", "let 0 := 123;")
	p.registerPrefix(tokNumber, parseNumber)
	stmt, err = parseDeclarationStmt(p)
	expectParserError(t, "(1:5) expected identifier", stmt, err)

	p = makeParser("", "let a = 123;")
	p.registerPrefix(tokNumber, parseNumber)
	stmt, err = parseDeclarationStmt(p)
	expectParserError(t, "(1:7) expected :=", stmt, err)

	p = makeParser("", "let a :=;")
	p.registerPrefix(tokNumber, parseNumber)
	stmt, err = parseDeclarationStmt(p)
	expectParserError(t, "(1:9) unexpected symbol", stmt, err)

	p = makeParser("", "let a := 123")
	p.registerPrefix(tokNumber, parseNumber)
	stmt, err = parseDeclarationStmt(p)
	expectParserError(t, "(1:12) expected semicolon", stmt, err)
}

func TestParseReturnStmt(t *testing.T) {
	p := makeParser("", "return;")
	stmt, err := parseReturnStmt(p)
	expectNoParserErrors(t, "(return)", stmt, err)
	expectStart(t, stmt, 1, 1)

	p = makeParser("", "return 123;")
	p.registerPrefix(tokNumber, parseNumber)
	stmt, err = parseReturnStmt(p)
	expectNoParserErrors(t, "(return 123)", stmt, err)

	p = makeParser("", "123;")
	stmt, err = parseReturnStmt(p)
	expectParserError(t, "(1:1) expected RETURN keyword", stmt, err)

	p = makeParser("", "return")
	stmt, err = parseReturnStmt(p)
	expectParserError(t, "(1:6) expected semicolon", stmt, err)

	p = makeParser("", "return let := 123;")
	p.registerPrefix(tokNumber, parseNumber)
	stmt, err = parseReturnStmt(p)
	expectParserError(t, "(1:8) unexpected symbol", stmt, err)
}

func TestParseExprStmt(t *testing.T) {
	p := makeParser("", "a := 123;")
	loadGrammar(p)
	stmt, err := parseExprStmt(p)
	expectNoParserErrors(t, "(= a 123)", stmt, err)
	expectStart(t, stmt, 1, 1)

	p = makeParser("", "callee(1, 2);")
	loadGrammar(p)
	stmt, err = parseExprStmt(p)
	expectNoParserErrors(t, "(callee (1 2))", stmt, err)
	expectStart(t, stmt, 1, 1)

	p = makeParser("", "a := 123")
	loadGrammar(p)
	stmt, err = parseExprStmt(p)
	expectParserError(t, "(1:8) expected semicolon", stmt, err)

	p = makeParser("", "let a := 123")
	loadGrammar(p)
	stmt, err = parseExprStmt(p)
	expectParserError(t, "(1:1) unexpected symbol", stmt, err)

	p = makeParser("", "2 + 2")
	loadGrammar(p)
	stmt, err = parseExprStmt(p)
	expectParserError(t, "(1:1) expected start of statement", stmt, err)
}

func TestParseTypeSig(t *testing.T) {
	expectTypeNote(t, parseTypeNote, "Int", "Int")
	expectTypeNote(t, parseTypeNote, "[Int]", "[Int]")
	expectTypeNote(t, parseTypeNote, "Int?", "Int?")
	expectTypeNote(t, parseTypeNote, "Int??", "Int??")
	expectTypeNote(t, parseTypeNote, "Int???", "Int???")
	expectTypeNote(t, parseTypeNote, "[Int?]", "[Int?]")
	expectTypeNote(t, parseTypeNote, "[Int?]?", "[Int?]?")
	expectTypeNote(t, parseTypeNote, "[Int]?", "[Int]?")
	expectTypeNote(t, parseTypeNote, "([Int]?, Bool)", "([Int]? Bool)")
	expectTypeNote(t, parseTypeNote, "() => [Int]?", "() => [Int]?")
	expectTypeNote(t, parseTypeNote, "() => Void", "() => Void")

	expectTypeNoteError(t, parseTypeNote, "[?]", "(1:2) unexpected symbol")
	expectTypeNoteError(t, parseTypeNote, `[@]`, "(1:2) unexpected symbol")
	expectTypeNoteError(t, parseTypeNote, "[Int", "(1:4) expected right bracket")
	expectTypeNoteError(t, parseTypeNote, "?", "(1:1) unexpected symbol")
}

func TestParseTypeIdent(t *testing.T) {
	expectTypeNote(t, parseTypeNoteIdent, "Int", "Int")
	expectTypeNote(t, parseTypeNoteIdent, "Any", "Any")
	expectTypeNoteError(t, parseTypeNoteIdent, "123", "(1:1) expected identifier")
}

func TestParseTypeList(t *testing.T) {
	expectTypeNote(t, parseTypeNoteList, "[Int]", "[Int]")
	expectTypeNoteError(t, parseTypeNoteList, "Int]", "(1:1) expected left bracket")
	expectTypeNoteError(t, parseTypeNoteList, "[?]", "(1:2) unexpected symbol")
}

func TestParseTypeOptional(t *testing.T) {
	expectTypeOpt := func(fn typeNoteParser, source string, ast string) {
		p := makeParser("", source)
		loadGrammar(p)
		sig, err := fn(p)
		expectNoParserErrors(t, sig.String(), sig, err)
		sig, err = parseTypeNoteOptional(p, sig)
		expectNoParserErrors(t, ast, sig, err)
	}

	expectTypeOptError := func(fn typeNoteParser, source string, msg string) {
		p := makeParser("", source)
		loadGrammar(p)
		sig, err := fn(p)
		expectNoParserErrors(t, sig.String(), sig, err)
		sig, err = parseTypeNoteOptional(p, sig)
		expectParserError(t, msg, sig, err)
	}

	expectTypeOpt(parseTypeNoteIdent, "Int?", "Int?")
	expectTypeOpt(parseTypeNoteList, "[Int]?", "[Int]?")

	expectTypeOptError(parseTypeNoteIdent, "Int", "(1:3) expected question mark")
}

func TestParseTypeTuple(t *testing.T) {
	expectTypeNote(t, parseTypeNoteTuple, "()", "()")
	expectTypeNote(t, parseTypeNoteTuple, "(Int)", "(Int)")
	expectTypeNote(t, parseTypeNoteTuple, "(Int, Int)", "(Int Int)")
	expectTypeNote(t, parseTypeNoteTuple, "(Int, Int,)", "(Int Int)")
	expectTypeNoteError(t, parseTypeNoteTuple, "Int)", "(1:1) expected left paren")
	expectTypeNoteError(t, parseTypeNoteTuple, "(123)", "(1:2) unexpected symbol")
	expectTypeNoteError(t, parseTypeNoteTuple, "(Int", "(1:4) expected right paren")
}

func TestParseTypeFunction(t *testing.T) {
	expectTypeNote(t, parseTypeNoteTuple, "()=>Nil", "() => Nil")
	expectTypeNote(t, parseTypeNoteTuple, "(a, b, c)=>[Int]", "(a b c) => [Int]")
	expectTypeNote(t, parseTypeNoteTuple, "(a, b, c,)=>[Int]", "(a b c) => [Int]")
	expectTypeNoteError(t, parseTypeNoteTuple, "() => 123", "(1:7) unexpected symbol")

	p := makeParser("", "= > Int")
	tuple := TypeNoteTuple{nop, []TypeNote{}}
	sig, err := parseTypeNoteFunction(p, tuple)
	expectParserError(t, "(1:1) expected arrow", sig, err)
}

func TestParseExpr(t *testing.T) {
	p := makeParser("", "+a")
	p.registerPrefix(tokPlus, parsePrefix)
	p.registerPrefix(tokIdent, parseIdent)
	expr, err := parseExpr(p, precLowest)
	expectNoParserErrors(t, "(+ a)", expr, err)

	p = makeParser("", "+")
	p.registerPrefix(tokPlus, parsePrefix)
	p.registerPrefix(tokIdent, parseIdent)
	expr, err = parseExpr(p, precLowest)
	expectParserError(t, "(1:1) unexpected symbol", expr, err)

	p = makeParser("", "a + b + c")
	p.registerPostfix(tokPlus, parseInfix, precSum)
	p.registerPrefix(tokIdent, parseIdent)
	expr, err = parseExpr(p, precLowest)
	expectNoParserErrors(t, "(+ (+ a b) c)", expr, err)

	p = makeParser("", "a + b * c")
	p.registerPostfix(tokPlus, parseInfix, precSum)
	p.registerPostfix(tokStar, parseInfix, precProduct)
	p.registerPrefix(tokIdent, parseIdent)
	expr, err = parseExpr(p, precLowest)
	expectNoParserErrors(t, "(+ a (* b c))", expr, err)

	p = makeParser("", "a +")
	p.registerPostfix(tokPlus, parseInfix, precSum)
	p.registerPrefix(tokIdent, parseIdent)
	expr, err = parseExpr(p, precLowest)
	expectParserError(t, "(1:3) unexpected symbol", expr, err)
}

func TestParseFunction(t *testing.T) {
	p := makeParser("", "fn ():Void {}")
	expr, err := parseFunction(p)
	expectNoParserErrors(t, "(fn ():Void {})", expr, err)
	expectStart(t, expr, 1, 1)

	p = makeParser("", "fn ():Void {}")
	expr, err = parseFunction(p)
	expectNoParserErrors(t, "(fn ():Void {})", expr, err)

	p = makeParser("", "fn ():Int {}")
	expr, err = parseFunction(p)
	expectNoParserErrors(t, "(fn ():Int {})", expr, err)

	p = makeParser("", "fn ():[Int?]? {}")
	expr, err = parseFunction(p)
	expectNoParserErrors(t, "(fn ():[Int?]? {})", expr, err)

	p = makeParser("", "fn (a:Int):Void { let x := 123; }")
	loadGrammar(p)
	expr, err = parseFunction(p)
	expectNoParserErrors(t, "(fn (a:Int):Void {\n  (let x 123)})", expr, err)

	p = makeParser("", "func (a) { let x := 123; }")
	expr, err = parseFunction(p)
	expectParserError(t, "(1:1) expected FN keyword", expr, err)

	p = makeParser("", "fn (,) { let x := 123; }")
	expr, err = parseFunction(p)
	expectParserError(t, "(1:5) expected identifier", expr, err)

	p = makeParser("", "fn (): { let x := 123; }")
	expr, err = parseFunction(p)
	expectParserError(t, "(1:8) unexpected symbol", expr, err)

	p = makeParser("", "fn ():Void { let x = 123; }")
	expr, err = parseFunction(p)
	expectParserError(t, "(1:20) expected :=", expr, err)
}

func TestParseFunctionParams(t *testing.T) {
	expectParams := func(prog string, exp string) {
		p := makeParser("", prog)
		params, err := parseFunctionParams(p)

		if err != nil {
			t.Fatalf("Expected no errors, got '%s'\n", err)
		}

		got := "("
		for _, param := range params {
			got += " " + param.String()
		}
		got += " )"
		if exp != got {
			t.Errorf("Expected %s, got %s\n", exp, got)
		}
	}

	expectParamError := func(prog string, msg string) {
		p := makeParser("", prog)
		params, err := parseFunctionParams(p)

		if err == nil {
			got := "("
			for _, param := range params {
				got += " " + param.String()
			}
			got += " )"
			t.Errorf("Expected an error, got %s\n", got)
		} else if err.Error() != msg {
			t.Errorf("Expected '%s', got '%s'\n", msg, err)
		}
	}

	expectParams("()", "( )")
	expectParams("(a : Int)", "( a:Int )")
	expectParams("(a : Int,)", "( a:Int )")
	expectParams("(a : Int,b:Bool)", "( a:Int b:Bool )")
	expectParams("(a : Int, b:Bool)", "( a:Int b:Bool )")
	expectParams("(a : Int, b:Bool?)", "( a:Int b:Bool? )")
	expectParams("(a : [Int]?, b:Bool?)", "( a:[Int]? b:Bool? )")

	expectParamError("(,)", "(1:2) expected identifier")
	expectParamError("(123)", "(1:2) expected identifier")
	expectParamError("(a:Int,,)", "(1:8) expected identifier")
	expectParamError("(a)", "(1:3) expected colon between parameter name and type")
	expectParamError("a:Int)", "(1:1) expected left paren")
	expectParamError("(a:Int", "(1:6) expected right paren")
}

func TestParseFunctionParam(t *testing.T) {
	p := makeParser("", "a:Int")
	param, err := parseFunctionParam(p)
	expectNoParserErrors(t, "a:Int", param, err)
	expectStart(t, param, 1, 1)

	p = makeParser("", "0:Int")
	param, err = parseFunctionParam(p)
	expectParserError(t, "(1:1) expected identifier", param, err)

	p = makeParser("", "a:456")
	param, err = parseFunctionParam(p)
	expectParserError(t, "(1:3) unexpected symbol", param, err)
}

func TestParseFunctionReturnSig(t *testing.T) {
	p := makeParser("", ": Int")
	sig, err := parseFunctionReturnSig(p)
	expectNoParserErrors(t, "Int", sig, err)

	p = makeParser("", "Void")
	sig, err = parseFunctionReturnSig(p)
	expectParserError(t, "(1:1) expected colon between parameters and return type", sig, err)

	p = makeParser("", ": 456")
	sig, err = parseFunctionReturnSig(p)
	expectParserError(t, "(1:3) unexpected symbol", sig, err)
}

func TestParseInfix(t *testing.T) {
	parser := makeParser("", "a + b")
	parser.registerPostfix(tokPlus, parseInfix, precSum)
	parser.registerPrefix(tokIdent, parseIdent)

	var left Expr
	var expr Expr
	var err error

	if left, err = parseIdent(parser); err != nil {
		t.Fatalf("Expected no errors, got %v\n", err)
	}

	expr, err = parseInfix(parser, left)
	expectNoParserErrors(t, "(+ a b)", expr, err)
	expectStart(t, expr, 1, 1)

	parser = makeParser("", "a +")
	parser.registerPostfix(tokPlus, parseInfix, precSum)
	parser.registerPrefix(tokIdent, parseIdent)

	if left, err = parseIdent(parser); err != nil {
		t.Fatalf("Expected no errors, got %v\n", err)
	}

	expr, err = parseInfix(parser, left)
	expectParserError(t, "(1:3) unexpected symbol", expr, err)
}

func TestParseList(t *testing.T) {
	good := func(source string, exp string) {
		p := makeParser("", source)
		loadGrammar(p)
		expr, err := parseList(p)
		expectNoParserErrors(t, exp, expr, err)
		expectStart(t, expr, 1, 1)
	}

	bad := func(source string, exp string) {
		p := makeParser("", source)
		loadGrammar(p)
		expr, err := parseList(p)
		expectParserError(t, exp, expr, err)
	}

	good("[]", "[ ]")
	good("[a]", "[ a ]")
	good("[a,]", "[ a ]")
	good("[a,b]", "[ a b ]")
	good("[ a, b, c]", "[ a b c ]")

	bad("a, b]", "(1:1) expected left bracket")
	bad("[ let ]", "(1:3) unexpected symbol")
	bad("[a, b", "(1:5) expected right bracket")
}

func TestParseSubscript(t *testing.T) {
	p := makeParser("", "abc[0]")
	loadGrammar(p)
	ident, _ := parseIdent(p)
	expr, err := parseSubscript(p, ident)
	expectNoParserErrors(t, "abc[0]", expr, err)
	expectStart(t, expr, 1, 1)

	p = makeParser("", "abc]")
	loadGrammar(p)
	ident, _ = parseIdent(p)
	expr, err = parseSubscript(p, ident)
	expectParserError(t, "(1:4) expect left bracket", expr, err)

	p = makeParser("", "abc[]")
	loadGrammar(p)
	ident, _ = parseIdent(p)
	expr, err = parseSubscript(p, ident)
	expectParserError(t, "(1:5) expected index expression", expr, err)

	p = makeParser("", "abc[let]")
	loadGrammar(p)
	ident, _ = parseIdent(p)
	expr, err = parseSubscript(p, ident)
	expectParserError(t, "(1:5) unexpected symbol", expr, err)

	p = makeParser("", "abc[0")
	loadGrammar(p)
	ident, _ = parseIdent(p)
	expr, err = parseSubscript(p, ident)
	expectParserError(t, "(1:5) expect right bracket", expr, err)
}

func TestParseAccessExpr(t *testing.T) {
	good := func(source string, exp string) {
		p := makeParser("", source)
		loadGrammar(p)
		expr, err := parseExpr(p, precLowest)
		expectNoParserErrors(t, exp, expr, err)
	}

	bad := func(source string, msg string) {
		p := makeParser("", source)
		loadGrammar(p)
		left, _ := parseExpr(p, precDispatch)
		expr, err := parseAccess(p, left)
		expectParserError(t, msg, expr, err)
	}

	good("a.b", "(a).b")
	good("a.b.c", "((a).b).c")
	good("a + b.c", "(+ a (b).c)")
	good("a.b + c", "(+ (a).b c)")
	good("a.b(c)", "((a).b (c))")
	good("a(b).c", "((a (b))).c")
	good("a[b].c", "(a[b]).c")
	good("a.b[c]", "(a).b[c]")

	bad("a b", "(1:3) expect dot")
	bad("a.", "(1:2) unexpected symbol")
}

func TestParseDispatchExpr(t *testing.T) {
	p := makeParser("", "callee()")
	loadGrammar(p)
	ident, _ := parseIdent(p)
	expr, err := parseDispatch(p, ident)
	expectNoParserErrors(t, "(callee ())", expr, err)
	expectStart(t, expr, 1, 1)

	p = makeParser("", "callee(1, 2, 3)")
	loadGrammar(p)
	expr, err = parseExpr(p, precLowest)
	expectNoParserErrors(t, "(callee (1 2 3))", expr, err)

	p = makeParser("", "callee)")
	loadGrammar(p)
	ident, _ = parseIdent(p)
	expr, err = parseDispatch(p, ident)
	expectParserError(t, "(1:7) expected left paren", expr, err)

	p = makeParser("", "callee(let")
	loadGrammar(p)
	ident, _ = parseIdent(p)
	expr, err = parseDispatch(p, ident)
	expectParserError(t, "(1:8) unexpected symbol", expr, err)

	p = makeParser("", "callee(123")
	loadGrammar(p)
	ident, _ = parseIdent(p)
	expr, err = parseDispatch(p, ident)
	expectParserError(t, "(1:10) expected right paren", expr, err)
}

func TestParseAssignExpr(t *testing.T) {
	p := makeParser("", "a := 123")
	loadGrammar(p)
	expr, err := parseExpr(p, precLowest)
	expectNoParserErrors(t, "(= a 123)", expr, err)
	expectStart(t, expr, 1, 1)

	p = makeParser("", "foo() := 123")
	loadGrammar(p)
	expr, err = parseExpr(p, precLowest)
	expectParserError(t, "(1:1) left hand must be an identifier", expr, err)

	p = makeParser("", "a :=")
	loadGrammar(p)
	expr, err = parseExpr(p, precLowest)
	expectParserError(t, "(1:4) unexpected symbol", expr, err)
}

func TestParsePostfix(t *testing.T) {
	parser := makeParser("", "a+")
	parser.registerPostfix(tokPlus, parsePostfix, precPostfix)
	parser.registerPrefix(tokIdent, parseIdent)

	var left Expr
	var expr Expr
	var err error

	if left, err = parseIdent(parser); err != nil {
		t.Fatalf("Expected no errors, got %v\n", err)
	}

	expr, err = parsePostfix(parser, left)
	expectNoParserErrors(t, "(+ a)", expr, err)
	expectStart(t, expr, 1, 1)
}

func TestParsePrefix(t *testing.T) {
	parser := makeParser("", "+a")
	parser.registerPrefix(tokPlus, parsePrefix)
	parser.registerPrefix(tokIdent, parseIdent)

	expr, err := parsePrefix(parser)
	expectNoParserErrors(t, "(+ a)", expr, err)
	expectStart(t, expr, 1, 1)

	parser = makeParser("", "+")
	parser.registerPrefix(tokPlus, parsePrefix)
	parser.registerPrefix(tokIdent, parseIdent)

	expr, err = parsePrefix(parser)
	expectParserError(t, "(1:1) unexpected symbol", expr, err)
}

func TestParseGroup(t *testing.T) {
	parser := makeParser("", "(a)")
	parser.registerPrefix(tokParenL, parsePrefix)
	parser.registerPrefix(tokIdent, parseIdent)

	expr, err := parseGroup(parser)
	expectNoParserErrors(t, "a", expr, err)
	expectStart(t, expr, 1, 2)

	parser = makeParser("", "a)")
	parser.registerPrefix(tokParenL, parsePrefix)
	parser.registerPrefix(tokIdent, parseIdent)

	expr, err = parseGroup(parser)
	expectParserError(t, "(1:1) expected left paren", expr, err)

	parser = makeParser("", "(")
	parser.registerPrefix(tokParenL, parsePrefix)
	parser.registerPrefix(tokIdent, parseIdent)

	expr, err = parseGroup(parser)
	expectParserError(t, "(1:1) unexpected symbol", expr, err)

	parser = makeParser("", "(a")
	parser.registerPrefix(tokParenL, parsePrefix)
	parser.registerPrefix(tokIdent, parseIdent)

	expr, err = parseGroup(parser)
	expectParserError(t, "(1:2) expected right paren", expr, err)
}

func TestParseSelf(t *testing.T) {
	parser := makeParser("", "self")

	expr, err := parseSelf(parser)
	expectNoParserErrors(t, "self", expr, err)
	expectStart(t, expr, 1, 1)

	parser = makeParser("", "selfx")

	expr, err = parseSelf(parser)
	expectParserError(t, "(1:1) expected self", expr, err)
}

func TestParseIdent(t *testing.T) {
	parser := makeParser("", "abc")

	expr, err := parseIdent(parser)
	expectNoParserErrors(t, "abc", expr, err)
	expectStart(t, expr, 1, 1)

	parser = makeParser("", "123")

	expr, err = parseIdent(parser)
	expectParserError(t, "(1:1) expected identifier", expr, err)
}

func TestParseNumber(t *testing.T) {
	parser := makeParser("", "123")

	expr, err := parseNumber(parser)
	expectNoParserErrors(t, "123", expr, err)
	expectStart(t, expr, 1, 1)

	parser = makeParser("", "abc")

	expr, err = parseNumber(parser)
	expectParserError(t, "(1:1) expected number literal", expr, err)

	loc := Loc{Line: 1, Col: 1}
	expr, err = evalNumber(parser, token{Type: tokNumber, Lexeme: "abc", Loc: loc})
	expectParserError(t, "(1:1) malformed number literal", expr, err)
}

func TestParseString(t *testing.T) {
	p := makeParser("", `"foo"`)
	expr, err := parseString(p)
	expectNoParserErrors(t, `"foo"`, expr, err)
	expectStart(t, expr, 1, 1)

	p = makeParser("", "123")
	expr, err = parseString(p)
	expectParserError(t, "(1:1) expected string literal", expr, err)

	p = makeParser("", "\"foo\n\"")
	expr, err = parseExpr(p, precLowest)
	expectParserError(t, "(1:5) unclosed string", expr, err)

	p = makeParser("", "\"foo")
	expr, err = parseExpr(p, precLowest)
	expectParserError(t, "(1:4) unclosed string", expr, err)
}

func TestParseBoolean(t *testing.T) {
	p := makeParser("", "true")
	expr, err := parseBoolean(p)
	expectNoParserErrors(t, "true", expr, err)
	expectStart(t, expr, 1, 1)

	p = makeParser("", "false")
	expr, err = parseBoolean(p)
	expectNoParserErrors(t, "false", expr, err)
	expectStart(t, expr, 1, 1)

	p = makeParser("", "flase")
	expr, err = parseBoolean(p)
	expectParserError(t, "(1:1) expected boolean literal", expr, err)

	loc := Loc{Line: 1, Col: 1}
	expr, err = evalBoolean(p, token{Type: tokBoolean, Lexeme: "ture", Loc: loc})
	expectParserError(t, "(1:1) malformed boolean literal", expr, err)
}

type typeNoteParser func(p *parser) (TypeNote, error)

func expectTypeNote(t *testing.T, fn typeNoteParser, source string, ast string) {
	t.Helper()
	p := makeParser("", source)
	loadGrammar(p)
	sig, err := fn(p)
	expectNoParserErrors(t, ast, sig, err)
	expectStart(t, sig, 1, 1)
}

func expectTypeNoteError(t *testing.T, fn typeNoteParser, source string, msg string) {
	t.Helper()
	p := makeParser("", source)
	loadGrammar(p)
	sig, err := fn(p)
	expectParserError(t, msg, sig, err)
}

func expectNoParserErrors(t *testing.T, ast string, node ASTNode, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no errors, got '%s'\n", err)
	} else {
		expectAST(t, ast, node)
	}
}

func expectParserError(t *testing.T, msg string, node ASTNode, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("Expected an error, got %s\n", node)
	} else if err.Error() != msg {
		t.Errorf("Expected '%s', got '%s'\n", msg, err)
	}
}

func expectAST(t *testing.T, ast string, got ASTNode) {
	t.Helper()
	if ast != got.String() {
		t.Errorf("Expected '%s', got '%s'\n", ast, got)
	}
}

func expectStart(t *testing.T, node ASTNode, line int, col int) {
	t.Helper()
	got := node.Start()
	exp := Loc{Line: line, Col: col}

	if exp.String() != got.String() {
		t.Errorf("Expected %s, got %s\n", exp, got)
	}
}
