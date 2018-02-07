package lang

import (
	"fmt"
	"strconv"
	"strings"
)

// ASTNode is the ancestor of all AST nodes
type ASTNode interface {
	Start() Loc
	String() string
	isNode()
}

// AST describes all top-level statements within a script
type AST struct {
	Stmts []Stmt
}

// Start returns a location that this node can be considered to start at
func (r AST) Start() Loc {
	if len(r.Stmts) > 0 {
		return r.Stmts[0].Start()
	}

	return Loc{Line: 1, Col: 1}
}

func (r AST) String() string {
	out := ""
	for i, stmt := range r.Stmts {
		if i > 0 {
			out += "\n"
		}

		out += stmt.String()
	}
	return out
}

func (r AST) isNode() {}

// Stmt describes all constructs that return no value
type Stmt interface {
	Start() Loc
	String() string
	isNode()
	isStmt()
}

// StmtBlock describes any series of statements bounded by curly braces
type StmtBlock struct {
	Left  token
	Stmts []Stmt
	Right token
}

// Start returns a location that this node can be considered to start at
func (sb StmtBlock) Start() Loc { return sb.Left.Loc }

func (sb StmtBlock) String() string {
	out := "{"
	for _, stmt := range sb.Stmts {
		out += "\n" + indentBlock("  ", stmt.String())
	}
	return out + "}"
}

func (sb StmtBlock) isNode() {}

// UseStmt describes a file or module import
type UseStmt struct {
	Tok    token
	Path   *StringExpr
	Filter []*UseFilter
}

// Start returns a location that this node can be considered to start at
func (s UseStmt) Start() Loc { return s.Tok.Loc }
func (s UseStmt) String() string {
	var filter string
	if len(s.Filter) > 0 {
		filter = " ("
		for i, f := range s.Filter {
			if i > 0 {
				filter += " "
			}
			filter += f.String()
		}
		filter += ")"
	}

	return fmt.Sprintf("(use %s%s)", s.Path, filter)
}
func (s UseStmt) isNode() {}
func (s UseStmt) isStmt() {}

// UseFilter describes a named import
type UseFilter struct {
	Name *IdentExpr
}

// Start returns a location that this node can be considered to start at
func (s UseFilter) Start() Loc     { return s.Name.Start() }
func (s UseFilter) String() string { return s.Name.String() }
func (s UseFilter) isNode()        {}

// PubStmt describes a file or module import
type PubStmt struct {
	Tok  token
	Stmt *DeclarationStmt
}

// Start returns a location that this node can be considered to start at
func (s PubStmt) Start() Loc     { return s.Tok.Loc }
func (s PubStmt) String() string { return fmt.Sprintf("(pub %s)", s.Stmt) }
func (s PubStmt) isNode()        {}
func (s PubStmt) isStmt()        {}

// IfStmt describes a condition expression and an associated clause
type IfStmt struct {
	Tok    token
	Cond   Expr
	Clause *StmtBlock
}

// Start returns a location that this node can be considered to start at
func (is IfStmt) Start() Loc     { return is.Tok.Loc }
func (is IfStmt) String() string { return fmt.Sprintf("(if %s %s)", is.Cond, is.Clause) }
func (is IfStmt) isNode()        {}
func (is IfStmt) isStmt()        {}

// DeclarationStmt describes the declaration and assignment of a variable
type DeclarationStmt struct {
	Tok  token
	Name *IdentExpr
	Expr Expr
}

// Start returns a location that this node can be considered to start at
func (ds DeclarationStmt) Start() Loc     { return ds.Tok.Loc }
func (ds DeclarationStmt) String() string { return fmt.Sprintf("(let %s %s)", ds.Name, ds.Expr) }
func (ds DeclarationStmt) isNode()        {}
func (ds DeclarationStmt) isStmt()        {}

// ReturnStmt describes a return keyword and an optional returned expression.
type ReturnStmt struct {
	Tok  token
	Expr Expr
}

// Start returns a location that this node can be considered to start at
func (rs ReturnStmt) Start() Loc { return rs.Tok.Loc }

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
func (es ExprStmt) Start() Loc     { return es.Expr.Start() }
func (es ExprStmt) String() string { return es.Expr.String() }
func (es ExprStmt) isNode()        {}
func (es ExprStmt) isStmt()        {}

// TypeNote describes a syntax type annotation
type TypeNote interface {
	Start() Loc
	String() string
	isNode()
	isType()
}

// TypeNoteAny describes a type that matches everything
type TypeNoteAny struct {
	Tok token
}

// Start returns a location that this node can be considered to start at
func (ta TypeNoteAny) Start() Loc     { return ta.Tok.Loc }
func (ta TypeNoteAny) String() string { return "Any" }
func (ta TypeNoteAny) isNode()        {}
func (ta TypeNoteAny) isType()        {}

// TypeNoteVoid describes a missing type annotation
type TypeNoteVoid struct {
	Tok token
}

// Start returns a location that this node can be considered to start at
func (tv TypeNoteVoid) Start() Loc     { return tv.Tok.Loc }
func (tv TypeNoteVoid) String() string { return "Void" }
func (tv TypeNoteVoid) isNode()        {}
func (tv TypeNoteVoid) isType()        {}

// TypeNoteTuple describes a set of 0 or more types wrapped in parentheses
type TypeNoteTuple struct {
	Tok   token
	Elems []TypeNote
}

// Start returns a location that this node can be considered to start at
func (tt TypeNoteTuple) Start() Loc { return tt.Tok.Loc }

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
func (tf TypeNoteFunction) Start() Loc { return tf.Params.Start() }

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
	Tok  token
	Name string
}

// Start returns a location that this node can be considered to start at
func (ti TypeNoteIdent) Start() Loc     { return ti.Tok.Loc }
func (ti TypeNoteIdent) String() string { return ti.Name }
func (ti TypeNoteIdent) isNode()        {}
func (ti TypeNoteIdent) isType()        {}

// TypeNoteList describes a list type
type TypeNoteList struct {
	Tok   token
	Child TypeNote
}

// Start returns a location that this node can be considered to start at
func (tl TypeNoteList) Start() Loc     { return tl.Tok.Loc }
func (tl TypeNoteList) String() string { return fmt.Sprintf("[%s]", tl.Child) }
func (tl TypeNoteList) isNode()        {}
func (tl TypeNoteList) isType()        {}

// TypeNoteOptional describes a list type
type TypeNoteOptional struct {
	Tok   token
	Child TypeNote
}

// Start returns a location that this node can be considered to start at
func (to TypeNoteOptional) Start() Loc     { return to.Child.Start() }
func (to TypeNoteOptional) String() string { return fmt.Sprintf("%s?", to.Child) }
func (to TypeNoteOptional) isNode()        {}
func (to TypeNoteOptional) isType()        {}

// Expr describes all constructs that resolve to a value
type Expr interface {
	Start() Loc
	String() string
	isNode()
	isExpr()
}

// FunctionExpr describes a function's entire type signature and body
type FunctionExpr struct {
	Tok    token
	Params []*FunctionParam
	Ret    TypeNote
	Block  *StmtBlock
}

// Start returns a location that this node can be considered to start at
func (fe FunctionExpr) Start() Loc { return fe.Tok.Loc }

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
	Name *IdentExpr
	Note TypeNote
}

// Start returns a location that this node can be considered to start at
func (fp FunctionParam) Start() Loc { return fp.Name.Start() }

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
func (de DispatchExpr) Start() Loc { return de.Callee.Start() }

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
	Tok   token
	Left  *IdentExpr
	Right Expr
}

// Start returns a location that this node can be considered to start at
func (ae AssignExpr) Start() Loc     { return ae.Left.Start() }
func (ae AssignExpr) String() string { return fmt.Sprintf("(= %s %s)", ae.Left, ae.Right) }
func (ae AssignExpr) isNode()        {}
func (ae AssignExpr) isExpr()        {}

// ListExpr describes a listeral list constructor
type ListExpr struct {
	Tok      token
	Elements []Expr
}

// Start returns a location that this node can be considered to start at
func (le ListExpr) Start() Loc { return le.Tok.Loc }
func (le ListExpr) String() string {
	out := "[ "
	for _, elem := range le.Elements {
		out += fmt.Sprintf("%s ", elem)
	}
	return out + "]"
}
func (le ListExpr) isNode() {}
func (le ListExpr) isExpr() {}

// SubscriptExpr describes an index-access on a list-like expression
type SubscriptExpr struct {
	ListLike Expr
	Index    Expr
}

// Start returns a location that this node can be considered to start at
func (se SubscriptExpr) Start() Loc     { return se.ListLike.Start() }
func (se SubscriptExpr) String() string { return fmt.Sprintf("%s[%s]", se.ListLike, se.Index) }
func (se SubscriptExpr) isNode()        {}
func (se SubscriptExpr) isExpr()        {}

// AccessExpr uses dot notation to retrieve a sub-object
type AccessExpr struct {
	Left  Expr
	Right Expr
}

// Start returns a location that this node can be considered to start at
func (e AccessExpr) Start() Loc     { return e.Left.Start() }
func (e AccessExpr) String() string { return fmt.Sprintf("(%s).%s", e.Left, e.Right) }
func (e AccessExpr) isNode()        {}
func (e AccessExpr) isExpr()        {}

// BinaryExpr describes any two expressions associated by an operator
type BinaryExpr struct {
	Oper  string
	Tok   token
	Left  Expr
	Right Expr
}

// Start returns a location that this node can be considered to start at
func (be BinaryExpr) Start() Loc     { return be.Left.Start() }
func (be BinaryExpr) String() string { return fmt.Sprintf("(%s %s %s)", be.Oper, be.Left, be.Right) }
func (be BinaryExpr) isNode()        {}
func (be BinaryExpr) isExpr()        {}

// UnaryExpr describes any single expression associated to an operator
type UnaryExpr struct {
	Oper string
	Tok  token
	Expr Expr
}

// Start returns a location that this node can be considered to start at
func (ue UnaryExpr) Start() Loc     { return smallerLoc(ue.Tok.Loc, ue.Expr.Start()) }
func (ue UnaryExpr) String() string { return fmt.Sprintf("(%s %s)", ue.Oper, ue.Expr) }
func (ue UnaryExpr) isNode()        {}
func (ue UnaryExpr) isExpr()        {}

// SelfExpr describes a reflexive reference to a function from within that function
type SelfExpr struct {
	Tok token
}

// Start returns a location that this node can be considered to start at
func (se SelfExpr) Start() Loc     { return se.Tok.Loc }
func (se SelfExpr) String() string { return "self" }
func (se SelfExpr) isNode()        {}
func (se SelfExpr) isExpr()        {}

// IdentExpr describes an identifier
type IdentExpr struct {
	Tok  token
	Name string
}

// Start returns a location that this node can be considered to start at
func (ie IdentExpr) Start() Loc     { return ie.Tok.Loc }
func (ie IdentExpr) String() string { return ie.Name }
func (ie IdentExpr) isNode()        {}
func (ie IdentExpr) isExpr()        {}

// StringExpr describes a string literal
type StringExpr struct {
	Tok token
	Val string
}

// Start returns a location that this node can be considered to start at
func (se StringExpr) Start() Loc     { return se.Tok.Loc }
func (se StringExpr) String() string { return fmt.Sprintf("\"%s\"", se.Val) }
func (se StringExpr) isNode()        {}
func (se StringExpr) isExpr()        {}

// NumberExpr describes a string literal
type NumberExpr struct {
	Tok token
	Val int
}

// Start returns a location that this node can be considered to start at
func (ne NumberExpr) Start() Loc     { return ne.Tok.Loc }
func (ne NumberExpr) String() string { return strconv.Itoa(ne.Val) }
func (ne NumberExpr) isNode()        {}
func (ne NumberExpr) isExpr()        {}

// BooleanExpr describes a boolean constant
type BooleanExpr struct {
	Tok token
	Val bool
}

// Start returns a location that this node can be considered to start at
func (be BooleanExpr) Start() Loc { return be.Tok.Loc }
func (be BooleanExpr) String() string {
	if be.Val {
		return "true"
	}

	return "false"
}
func (be BooleanExpr) isNode() {}
func (be BooleanExpr) isExpr() {}

func indentBlock(indent string, source string) string {
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}

	return strings.Join(lines, "\n")
}
