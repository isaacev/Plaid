package parser

import (
	"fmt"
	"plaid/lexer"
	"strconv"
	"strings"
)

// Node is the ancestor of all AST nodes
type Node interface {
	String() string
	isNode()
}

// Program describes all top-level statements within a script
type Program struct {
	stmts []Stmt
}

func (p Program) String() string {
	out := ""
	for i, stmt := range p.stmts {
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

func (ds DeclarationStmt) String() string { return fmt.Sprintf("(let %s %s)", ds.name, ds.expr) }
func (ds DeclarationStmt) isNode()        {}
func (ds DeclarationStmt) isStmt()        {}

// TypeSig describes a syntax type annotation
type TypeSig interface {
	String() string
	isNode()
	isType()
}

// TypeIdent describes a named reference to a type
type TypeIdent struct {
	tok  lexer.Token
	name string
}

func (ti TypeIdent) String() string { return ti.name }
func (ti TypeIdent) isNode()        {}
func (ti TypeIdent) isType()        {}

// TypeList describes a list type
type TypeList struct {
	tok   lexer.Token
	child TypeSig
}

func (tl TypeList) String() string { return fmt.Sprintf("[%s]", tl.child) }
func (tl TypeList) isNode()        {}
func (tl TypeList) isType()        {}

// TypeOptional describes a list type
type TypeOptional struct {
	tok   lexer.Token
	child TypeSig
}

func (to TypeOptional) String() string { return fmt.Sprintf("%s?", to.child) }
func (to TypeOptional) isNode()        {}
func (to TypeOptional) isType()        {}

// Expr describes all constructs that resolve to a value
type Expr interface {
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

func (fe FunctionExpr) String() string {
	out := "(fn ("
	for i, param := range fe.params {
		if i > 0 {
			out += " "
		}
		out += param.String()
	}
	out += fmt.Sprintf("):%s %s)", fe.ret, fe.block)
	return out
}

func (fe FunctionExpr) isExpr() {}
func (fe FunctionExpr) isNode() {}

// FunctionParam describes a single function argument's name and type signature
type FunctionParam struct {
	name IdentExpr
	sig  TypeSig
}

func (fp FunctionParam) String() string {
	if fp.sig != nil {
		return fmt.Sprintf("%s:%s", fp.name, fp.sig)
	}

	return fp.name.String()
}

// BinaryExpr describes any two expressions associated by an operator
type BinaryExpr struct {
	oper  string
	tok   lexer.Token
	left  Expr
	right Expr
}

func (be BinaryExpr) String() string { return fmt.Sprintf("(%s %s %s)", be.oper, be.left, be.right) }
func (be BinaryExpr) isNode()        {}
func (be BinaryExpr) isExpr()        {}

// UnaryExpr describes any single expression associated to an operator
type UnaryExpr struct {
	oper string
	tok  lexer.Token
	expr Expr
}

func (ue UnaryExpr) String() string { return fmt.Sprintf("(%s %s)", ue.oper, ue.expr) }
func (ue UnaryExpr) isNode()        {}
func (ue UnaryExpr) isExpr()        {}

// IdentExpr describes an identifier
type IdentExpr struct {
	tok  lexer.Token
	name string
}

func (ie IdentExpr) String() string { return ie.name }
func (ie IdentExpr) isNode()        {}
func (ie IdentExpr) isExpr()        {}

// StringExpr describes a string literal
type StringExpr struct {
	tok lexer.Token
	val string
}

func (se StringExpr) String() string { return fmt.Sprintf("\"%s\"", se.val) }
func (se StringExpr) isNode()        {}
func (se StringExpr) isExpr()        {}

// NumberExpr describes a string literal
type NumberExpr struct {
	tok lexer.Token
	val int
}

func (ne NumberExpr) String() string { return strconv.Itoa(ne.val) }
func (ne NumberExpr) isNode()        {}
func (ne NumberExpr) isExpr()        {}

func indentBlock(indent string, source string) string {
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}

	return strings.Join(lines, "\n")
}
