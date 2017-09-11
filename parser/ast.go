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
	Left  lexer.Token
	Stmts []Stmt
	Right lexer.Token
}

// Start returns a location that this node can be considered to start at
func (sb StmtBlock) Start() lexer.Loc { return sb.Left.Loc }

func (sb StmtBlock) String() string {
	out := "{"
	for _, stmt := range sb.Stmts {
		out += "\n" + indentBlock("  ", stmt.String())
	}
	return out + "}"
}

func (sb StmtBlock) isNode() {}

// DeclarationStmt describes the declaration and assignment of a variable
type DeclarationStmt struct {
	Tok  lexer.Token
	Name IdentExpr
	Expr Expr
}

// Start returns a location that this node can be considered to start at
func (ds DeclarationStmt) Start() lexer.Loc { return ds.Tok.Loc }
func (ds DeclarationStmt) String() string   { return fmt.Sprintf("(let %s %s)", ds.Name, ds.Expr) }
func (ds DeclarationStmt) isNode()          {}
func (ds DeclarationStmt) isStmt()          {}

// ReturnStmt describes a return keyword and an optional returned expression.
type ReturnStmt struct {
	Tok  lexer.Token
	Expr Expr
}

// Start returns a location that this node can be considered to start at
func (rs ReturnStmt) Start() lexer.Loc { return rs.Tok.Loc }

func (rs ReturnStmt) String() string {
	if rs.Expr != nil {
		return fmt.Sprintf("(return %s)", rs.Expr)
	}

	return "(return)"
}

func (rs ReturnStmt) isNode() {}
func (rs ReturnStmt) isStmt() {}

// ExprStmt describes certain expressions that can be used in the place of statements
type ExprStmt struct {
	Expr Expr
}

// Start returns a location that this node can be considered to start at
func (es ExprStmt) Start() lexer.Loc { return es.Expr.Start() }
func (es ExprStmt) String() string   { return es.Expr.String() }
func (es ExprStmt) isNode()          {}
func (es ExprStmt) isStmt()          {}

// TypeNote describes a syntax type annotation
type TypeNote interface {
	Start() lexer.Loc
	String() string
	isNode()
	isType()
}

// TypeNoteTuple describes a set of 0 or more types wrapped in parentheses
type TypeNoteTuple struct {
	Tok   lexer.Token
	Elems []TypeNote
}

// Start returns a location that this node can be considered to start at
func (tt TypeNoteTuple) Start() lexer.Loc { return tt.Tok.Loc }

func (tt TypeNoteTuple) String() string {
	out := "("
	for i, elem := range tt.Elems {
		if i > 0 {
			out += " "
		}
		out += elem.String()
	}
	out += ")"
	return out
}

func (tt TypeNoteTuple) isNode() {}
func (tt TypeNoteTuple) isType() {}

// TypeNoteFunction describes a function type annotation
type TypeNoteFunction struct {
	Params TypeNoteTuple
	Ret    TypeNote
}

// Start returns a location that this node can be considered to start at
func (tf TypeNoteFunction) Start() lexer.Loc { return tf.Params.Start() }

func (tf TypeNoteFunction) String() string {
	out := tf.Params.String()
	out += " => "
	out += tf.Ret.String()
	return out
}

func (tf TypeNoteFunction) isNode() {}
func (tf TypeNoteFunction) isType() {}

// TypeNoteIdent describes a named reference to a type
type TypeNoteIdent struct {
	Tok  lexer.Token
	Name string
}

// Start returns a location that this node can be considered to start at
func (ti TypeNoteIdent) Start() lexer.Loc { return ti.Tok.Loc }
func (ti TypeNoteIdent) String() string   { return ti.Name }
func (ti TypeNoteIdent) isNode()          {}
func (ti TypeNoteIdent) isType()          {}

// TypeNoteList describes a list type
type TypeNoteList struct {
	Tok   lexer.Token
	Child TypeNote
}

// Start returns a location that this node can be considered to start at
func (tl TypeNoteList) Start() lexer.Loc { return tl.Tok.Loc }
func (tl TypeNoteList) String() string   { return fmt.Sprintf("[%s]", tl.Child) }
func (tl TypeNoteList) isNode()          {}
func (tl TypeNoteList) isType()          {}

// TypeNoteOptional describes a list type
type TypeNoteOptional struct {
	Tok   lexer.Token
	Child TypeNote
}

// Start returns a location that this node can be considered to start at
func (to TypeNoteOptional) Start() lexer.Loc { return to.Child.Start() }
func (to TypeNoteOptional) String() string   { return fmt.Sprintf("%s?", to.Child) }
func (to TypeNoteOptional) isNode()          {}
func (to TypeNoteOptional) isType()          {}

// Expr describes all constructs that resolve to a value
type Expr interface {
	Start() lexer.Loc
	String() string
	isNode()
	isExpr()
}

// FunctionExpr describes a function's entire type signature and body
type FunctionExpr struct {
	Tok    lexer.Token
	Params []FunctionParam
	Ret    TypeNote
	Block  StmtBlock
}

// Start returns a location that this node can be considered to start at
func (fe FunctionExpr) Start() lexer.Loc { return fe.Tok.Loc }

func (fe FunctionExpr) String() string {
	out := "(fn ("
	for i, param := range fe.Params {
		if i > 0 {
			out += " "
		}
		out += param.String()
	}
	out += ")"
	if fe.Ret != nil {
		out += fmt.Sprintf(":%s", fe.Ret)
	}
	out += fmt.Sprintf(" %s)", fe.Block)
	return out
}

func (fe FunctionExpr) isExpr() {}
func (fe FunctionExpr) isNode() {}

// FunctionParam describes a single function argument's name and type annotation
type FunctionParam struct {
	Name IdentExpr
	Note TypeNote
}

// Start returns a location that this node can be considered to start at
func (fp FunctionParam) Start() lexer.Loc { return fp.Name.Start() }

func (fp FunctionParam) String() string {
	if fp.Note != nil {
		return fmt.Sprintf("%s:%s", fp.Name, fp.Note)
	}

	return fp.Name.String()
}

func (fp FunctionParam) isNode() {}

// DispatchExpr describes a function call including the callee and any arguments
type DispatchExpr struct {
	Callee Expr
	Args   []Expr
}

// Start returns a location that this node can be considered to start at
func (de DispatchExpr) Start() lexer.Loc { return de.Callee.Start() }

func (de DispatchExpr) String() string {
	out := "("
	out += de.Callee.String()
	out += " ("
	for i, arg := range de.Args {
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
	Tok   lexer.Token
	Left  IdentExpr
	Right Expr
}

// Start returns a location that this node can be considered to start at
func (ae AssignExpr) Start() lexer.Loc { return ae.Left.Start() }
func (ae AssignExpr) String() string   { return fmt.Sprintf("(= %s %s)", ae.Left, ae.Right) }
func (ae AssignExpr) isNode()          {}
func (ae AssignExpr) isExpr()          {}

// BinaryExpr describes any two expressions associated by an operator
type BinaryExpr struct {
	Oper  string
	Tok   lexer.Token
	Left  Expr
	Right Expr
}

// Start returns a location that this node can be considered to start at
func (be BinaryExpr) Start() lexer.Loc { return be.Left.Start() }
func (be BinaryExpr) String() string   { return fmt.Sprintf("(%s %s %s)", be.Oper, be.Left, be.Right) }
func (be BinaryExpr) isNode()          {}
func (be BinaryExpr) isExpr()          {}

// UnaryExpr describes any single expression associated to an operator
type UnaryExpr struct {
	Oper string
	Tok  lexer.Token
	Expr Expr
}

// Start returns a location that this node can be considered to start at
func (ue UnaryExpr) Start() lexer.Loc { return lexer.SmallerLoc(ue.Tok.Loc, ue.Expr.Start()) }
func (ue UnaryExpr) String() string   { return fmt.Sprintf("(%s %s)", ue.Oper, ue.Expr) }
func (ue UnaryExpr) isNode()          {}
func (ue UnaryExpr) isExpr()          {}

// IdentExpr describes an identifier
type IdentExpr struct {
	Tok  lexer.Token
	Name string
}

// Start returns a location that this node can be considered to start at
func (ie IdentExpr) Start() lexer.Loc { return ie.Tok.Loc }
func (ie IdentExpr) String() string   { return ie.Name }
func (ie IdentExpr) isNode()          {}
func (ie IdentExpr) isExpr()          {}

// StringExpr describes a string literal
type StringExpr struct {
	Tok lexer.Token
	Val string
}

// Start returns a location that this node can be considered to start at
func (se StringExpr) Start() lexer.Loc { return se.Tok.Loc }
func (se StringExpr) String() string   { return fmt.Sprintf("\"%s\"", se.Val) }
func (se StringExpr) isNode()          {}
func (se StringExpr) isExpr()          {}

// NumberExpr describes a string literal
type NumberExpr struct {
	Tok lexer.Token
	Val int
}

// Start returns a location that this node can be considered to start at
func (ne NumberExpr) Start() lexer.Loc { return ne.Tok.Loc }
func (ne NumberExpr) String() string   { return strconv.Itoa(ne.Val) }
func (ne NumberExpr) isNode()          {}
func (ne NumberExpr) isExpr()          {}

func indentBlock(indent string, source string) string {
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}

	return strings.Join(lines, "\n")
}
