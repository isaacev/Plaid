package parser

import (
	"fmt"
	"plaid/lexer"
	"strconv"
	"strings"
)

// Node is the ancestor of all AST nodes
type Node interface {
	Start() lexer.Loc
	String() string
	isNode()
}

// Program describes all top-level statements within a script
type Program struct {
	Stmts []Stmt
}

// Start returns a location that this node can be considered to start at
func (p Program) Start() lexer.Loc {
	if len(p.Stmts) > 0 {
		return p.Stmts[0].Start()
	}

	return lexer.Loc{Line: 1, Col: 1}
}

func (p Program) String() string {
	out := ""
	for i, stmt := range p.Stmts {
		if i > 0 {
			out += "\n"
		}

		out += stmt.String()
	}
	return out
}

func (p Program) isNode() {}

// Stmt describes all constructs that return no value
type Stmt interface {
	Start() lexer.Loc
	String() string
	isNode()
	isStmt()
}

// StmtBlock describes any series of statements bounded by curly braces
type StmtBlock struct {
	left  lexer.Token
	stmts []Stmt
	right lexer.Token
}

// Start returns a location that this node can be considered to start at
func (sb StmtBlock) Start() lexer.Loc { return sb.left.Loc }

func (sb StmtBlock) String() string {
	out := "{"
	for _, stmt := range sb.stmts {
		out += "\n" + indentBlock("  ", stmt.String())
	}
	return out + "}"
}

func (sb StmtBlock) isNode() {}

// DeclarationStmt describes the declaration and assignment of a variable
type DeclarationStmt struct {
	tok  lexer.Token
	name IdentExpr
	expr Expr
}

// Start returns a location that this node can be considered to start at
func (ds DeclarationStmt) Start() lexer.Loc { return ds.tok.Loc }
func (ds DeclarationStmt) String() string   { return fmt.Sprintf("(let %s %s)", ds.name, ds.expr) }
func (ds DeclarationStmt) isNode()          {}
func (ds DeclarationStmt) isStmt()          {}

// ReturnStmt describes a return keyword and an optional returned expression.
type ReturnStmt struct {
	tok  lexer.Token
	expr Expr
}

// Start returns a location that this node can be considered to start at
func (rs ReturnStmt) Start() lexer.Loc { return rs.tok.Loc }

func (rs ReturnStmt) String() string {
	if rs.expr != nil {
		return fmt.Sprintf("(return %s)", rs.expr)
	}

	return "(return)"
}

func (rs ReturnStmt) isNode() {}
func (rs ReturnStmt) isStmt() {}

// ExprStmt describes certain expressions that can be used in the place of statements
type ExprStmt struct {
	expr Expr
}

// Start returns a location that this node can be considered to start at
func (es ExprStmt) Start() lexer.Loc { return es.expr.Start() }
func (es ExprStmt) String() string   { return es.expr.String() }
func (es ExprStmt) isNode()          {}
func (es ExprStmt) isStmt()          {}

// TypeSig describes a syntax type annotation
type TypeSig interface {
	Start() lexer.Loc
	String() string
	isNode()
	isType()
}

// TypeTuple describes a set of 0 or more types wrapped in parentheses
type TypeTuple struct {
	tok   lexer.Token
	elems []TypeSig
}

// Start returns a location that this node can be considered to start at
func (tt TypeTuple) Start() lexer.Loc { return tt.tok.Loc }

func (tt TypeTuple) String() string {
	out := "("
	for i, elem := range tt.elems {
		if i > 0 {
			out += " "
		}
		out += elem.String()
	}
	out += ")"
	return out
}

func (tt TypeTuple) isNode() {}
func (tt TypeTuple) isType() {}

// TypeFunction describes a function type annotation
type TypeFunction struct {
	params TypeTuple
	ret    TypeSig
}

// Start returns a location that this node can be considered to start at
func (tf TypeFunction) Start() lexer.Loc { return tf.params.Start() }

func (tf TypeFunction) String() string {
	out := tf.params.String()
	out += " => "
	out += tf.ret.String()
	return out
}

func (tf TypeFunction) isNode() {}
func (tf TypeFunction) isType() {}

// TypeIdent describes a named reference to a type
type TypeIdent struct {
	tok  lexer.Token
	name string
}

// Start returns a location that this node can be considered to start at
func (ti TypeIdent) Start() lexer.Loc { return ti.tok.Loc }
func (ti TypeIdent) String() string   { return ti.name }
func (ti TypeIdent) isNode()          {}
func (ti TypeIdent) isType()          {}

// TypeList describes a list type
type TypeList struct {
	tok   lexer.Token
	child TypeSig
}

// Start returns a location that this node can be considered to start at
func (tl TypeList) Start() lexer.Loc { return tl.tok.Loc }
func (tl TypeList) String() string   { return fmt.Sprintf("[%s]", tl.child) }
func (tl TypeList) isNode()          {}
func (tl TypeList) isType()          {}

// TypeOptional describes a list type
type TypeOptional struct {
	tok   lexer.Token
	child TypeSig
}

// Start returns a location that this node can be considered to start at
func (to TypeOptional) Start() lexer.Loc { return to.child.Start() }
func (to TypeOptional) String() string   { return fmt.Sprintf("%s?", to.child) }
func (to TypeOptional) isNode()          {}
func (to TypeOptional) isType()          {}

// Expr describes all constructs that resolve to a value
type Expr interface {
	Start() lexer.Loc
	String() string
	isNode()
	isExpr()
}

// FunctionExpr describes a function's entire type signature and body
type FunctionExpr struct {
	tok    lexer.Token
	params []FunctionParam
	ret    TypeSig
	block  StmtBlock
}

// Start returns a location that this node can be considered to start at
func (fe FunctionExpr) Start() lexer.Loc { return fe.tok.Loc }

func (fe FunctionExpr) String() string {
	out := "(fn ("
	for i, param := range fe.params {
		if i > 0 {
			out += " "
		}
		out += param.String()
	}
	out += ")"
	if fe.ret != nil {
		out += fmt.Sprintf(":%s", fe.ret)
	}
	out += fmt.Sprintf(" %s)", fe.block)
	return out
}

func (fe FunctionExpr) isExpr() {}
func (fe FunctionExpr) isNode() {}

// FunctionParam describes a single function argument's name and type signature
type FunctionParam struct {
	name IdentExpr
	sig  TypeSig
}

// Start returns a location that this node can be considered to start at
func (fp FunctionParam) Start() lexer.Loc { return fp.name.Start() }

func (fp FunctionParam) String() string {
	if fp.sig != nil {
		return fmt.Sprintf("%s:%s", fp.name, fp.sig)
	}

	return fp.name.String()
}

func (fp FunctionParam) isNode() {}

// DispatchExpr describes a function call including the callee and any arguments
type DispatchExpr struct {
	callee Expr
	args   []Expr
}

// Start returns a location that this node can be considered to start at
func (de DispatchExpr) Start() lexer.Loc { return de.callee.Start() }

func (de DispatchExpr) String() string {
	out := "("
	out += de.callee.String()
	out += " ("
	for i, arg := range de.args {
		if i > 0 {
			out += " "
		}
		out += arg.String()
	}
	out += "))"
	return out
}

func (de DispatchExpr) isNode() {}
func (de DispatchExpr) isExpr() {}

// AssignExpr describes the binding of a value to an assignable expression
type AssignExpr struct {
	tok   lexer.Token
	left  Expr
	right Expr
}

// Start returns a location that this node can be considered to start at
func (ae AssignExpr) Start() lexer.Loc { return ae.left.Start() }
func (ae AssignExpr) String() string   { return fmt.Sprintf("(= %s %s)", ae.left, ae.right) }
func (ae AssignExpr) isNode()          {}
func (ae AssignExpr) isExpr()          {}

// BinaryExpr describes any two expressions associated by an operator
type BinaryExpr struct {
	oper  string
	tok   lexer.Token
	left  Expr
	right Expr
}

// Start returns a location that this node can be considered to start at
func (be BinaryExpr) Start() lexer.Loc { return be.left.Start() }
func (be BinaryExpr) String() string   { return fmt.Sprintf("(%s %s %s)", be.oper, be.left, be.right) }
func (be BinaryExpr) isNode()          {}
func (be BinaryExpr) isExpr()          {}

// UnaryExpr describes any single expression associated to an operator
type UnaryExpr struct {
	oper string
	tok  lexer.Token
	expr Expr
}

// Start returns a location that this node can be considered to start at
func (ue UnaryExpr) Start() lexer.Loc { return lexer.SmallerLoc(ue.tok.Loc, ue.expr.Start()) }
func (ue UnaryExpr) String() string   { return fmt.Sprintf("(%s %s)", ue.oper, ue.expr) }
func (ue UnaryExpr) isNode()          {}
func (ue UnaryExpr) isExpr()          {}

// IdentExpr describes an identifier
type IdentExpr struct {
	tok  lexer.Token
	name string
}

// Start returns a location that this node can be considered to start at
func (ie IdentExpr) Start() lexer.Loc { return ie.tok.Loc }
func (ie IdentExpr) String() string   { return ie.name }
func (ie IdentExpr) isNode()          {}
func (ie IdentExpr) isExpr()          {}

// StringExpr describes a string literal
type StringExpr struct {
	tok lexer.Token
	val string
}

// Start returns a location that this node can be considered to start at
func (se StringExpr) Start() lexer.Loc { return se.tok.Loc }
func (se StringExpr) String() string   { return fmt.Sprintf("\"%s\"", se.val) }
func (se StringExpr) isNode()          {}
func (se StringExpr) isExpr()          {}

// NumberExpr describes a string literal
type NumberExpr struct {
	tok lexer.Token
	val int
}

// Start returns a location that this node can be considered to start at
func (ne NumberExpr) Start() lexer.Loc { return ne.tok.Loc }
func (ne NumberExpr) String() string   { return strconv.Itoa(ne.val) }
func (ne NumberExpr) isNode()          {}
func (ne NumberExpr) isExpr()          {}

func indentBlock(indent string, source string) string {
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}

	return strings.Join(lines, "\n")
}
