package codegen

import (
	"fmt"
	"plaid/check"
	"plaid/types"
	"strings"
)

var varRecordID uint

// VarRecord represents a declared variable
type VarRecord struct {
	ID   uint
	Name string
	typ  types.Type
}

func (vr *VarRecord) String() string { return fmt.Sprintf("%s<%d>", vr.Name, vr.ID) }

// IR represents the root of the intermediate representation tree
type IR struct {
	Scope    *LexicalScope
	Children []IRVoidNode
}

func (ir *IR) String() string {
	out := "(local"
	for _, record := range ir.Scope.Local {
		out += fmt.Sprintf(" %s:%s", record, record.typ)
	}
	out += ")"
	for _, child := range ir.Children {
		out += fmt.Sprintf("\n%s", child.String())
	}
	return out
}

// IRVoidNode represents IR nodes that do not resolve to a value in the VM
type IRVoidNode interface {
	String() string
	isVoidNode()
}

// IRTypedNode represents IR nodes that are resolved to a value in the VM
type IRTypedNode interface {
	Type() types.Type
	String() string
	isTypedNode()
}

// IRPrintNode represents an expression that should be printed to stdout
type IRPrintNode struct {
	Child IRTypedNode
}

func (pn IRPrintNode) String() string { return fmt.Sprintf("(print %s)", pn.Child) }
func (pn IRPrintNode) isVoidNode()    {}

// IRReturnNode represents a function termination statement that optionally
// pushes a typed value onto the stack
type IRReturnNode struct {
	Child IRTypedNode
}

func (rn IRReturnNode) String() string {
	if rn.Child != nil {
		return fmt.Sprintf("(return %s)", rn.Child)
	}

	return "(return)"
}
func (rn IRReturnNode) isVoidNode() {}

// IRVoidedNode wraps an IRTypedNode and treats that node as if it returned nothing
type IRVoidedNode struct {
	Child IRTypedNode
}

func (vn IRVoidedNode) String() string { return vn.Child.String() }
func (vn IRVoidedNode) isVoidNode()    {}

// IRFunctionNode represents a function literal
type IRFunctionNode struct {
	Scope  *LexicalScope
	Params []*VarRecord
	Ret    types.Type
	Body   []IRVoidNode
}

// Type returns the value type that this node resolves to
func (fn IRFunctionNode) Type() types.Type { return fn.Ret }
func (fn IRFunctionNode) String() string {
	out := "(fn ("
	for i, param := range fn.Params {
		if i > 0 {
			out += " "
		}
		out += param.String()
	}
	out += ")"
	if fn.Ret != nil {
		out += fmt.Sprintf(":%s", fn.Ret)
	}
	out += " {"
	if len(fn.Scope.Local) > 0 {
		out += "\n  (local"
		for _, record := range fn.Scope.Local {
			out += fmt.Sprintf(" %s:%s", record, record.typ)
		}
		out += ")"
	}
	for _, node := range fn.Body {
		out += "\n" + indentBlock("  ", node.String())
	}
	out += "})"
	return out
}
func (fn IRFunctionNode) isTypedNode() {}

// IRDispatchNode represents a call to a function along with any arguments
type IRDispatchNode struct {
	Callee IRTypedNode
	Args   []IRTypedNode
}

// Type returns the value type that this node resolves to
func (dn IRDispatchNode) Type() types.Type { return dn.Callee.Type() }
func (dn IRDispatchNode) String() string {
	out := "("
	out += dn.Callee.String()
	out += " ("
	for i, arg := range dn.Args {
		if i > 0 {
			out += " "
		}
		out += arg.String()
	}
	out += "))"
	return out
}
func (dn IRDispatchNode) isTypedNode() {}

// IRAssignNode represents an IR node that changes the value of a variable
type IRAssignNode struct {
	Record *VarRecord
	Child  IRTypedNode
}

// Type returns the value type that this node resolves to
func (an IRAssignNode) Type() types.Type { return an.Child.Type() }
func (an IRAssignNode) String() string   { return fmt.Sprintf("(= %s %s)", an.Record, an.Child) }
func (an IRAssignNode) isTypedNode()     {}

// IRBinaryNode reprents a binary expression and operator
type IRBinaryNode struct {
	Oper  string
	Left  IRTypedNode
	Right IRTypedNode
}

// Type returns the value type that this node resolves to
func (bn IRBinaryNode) Type() types.Type { return check.BuiltinInt }
func (bn IRBinaryNode) String() string   { return fmt.Sprintf("(%s %s %s)", bn.Oper, bn.Left, bn.Right) }
func (bn IRBinaryNode) isTypedNode()     {}

// IRReferenceNode resolves to a known variable
type IRReferenceNode struct {
	Record *VarRecord
}

// Type returns the value type that this node resolves to
func (rn IRReferenceNode) Type() types.Type { return check.BuiltinInt }
func (rn IRReferenceNode) String() string   { return rn.Record.String() }
func (rn IRReferenceNode) isTypedNode()     {}

// IRIntegerLiteralNode represents an integer value
type IRIntegerLiteralNode struct {
	Val int64
}

// Type returns the value type that this node resolves to
func (iln IRIntegerLiteralNode) Type() types.Type { return check.BuiltinInt }
func (iln IRIntegerLiteralNode) String() string   { return fmt.Sprintf("%d", iln.Val) }
func (iln IRIntegerLiteralNode) isTypedNode()     {}

// IRStringLiteralNode represents a string value
type IRStringLiteralNode struct {
	Val string
}

// Type returns the value type that this node resolves to
func (sln IRStringLiteralNode) Type() types.Type { return check.BuiltinStr }
func (sln IRStringLiteralNode) String() string   { return sln.Val }
func (sln IRStringLiteralNode) isTypedNode()     {}

func indentBlock(indent string, source string) string {
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}

	return strings.Join(lines, "\n")
}
