package codegen

import (
	"fmt"
	"plaid/parser"
	"plaid/scope"
	"plaid/vm"
)

// Transform converts an AST to the intermediate representation (IR) which is a
// precursor to the compiled bytecode
func Transform(prog *parser.Program, modules ...*vm.Module) IR {
	global := prog.Scope
	nodes := transformStmts(global, prog.Stmts)
	return IR{global, nodes}
}

func transformStmts(s scope.Scope, stmts []parser.Stmt) []IRVoidNode {
	var nodes []IRVoidNode
	for _, stmt := range stmts {
		node := transformStmt(s, stmt)
		nodes = append(nodes, node)
	}
	return nodes
}

func transformStmt(s scope.Scope, stmt parser.Stmt) IRVoidNode {
	switch stmt := stmt.(type) {
	case *parser.IfStmt:
		return transformIfStmt(s, stmt)
	case *parser.DeclarationStmt:
		return transformDeclarationStmt(s, stmt)
	case *parser.ReturnStmt:
		return transformReturnStmt(s, stmt)
	case *parser.ExprStmt:
		return transformExprStmt(s, stmt)
	default:
		panic(fmt.Sprintf("cannot transform %T", stmt))
	}
}

func transformIfStmt(s scope.Scope, stmt *parser.IfStmt) IRVoidNode {
	cond := transformExpr(s, stmt.Cond)
	clause := transformStmts(s, stmt.Clause.Stmts)
	return IRCondNode{cond, clause}
}

func transformDeclarationStmt(s scope.Scope, stmt *parser.DeclarationStmt) IRVoidNode {
	name := stmt.Name.Name
	child := transformExpr(s, stmt.Expr)
	reg := s.GetVariableRegister(name)
	return IRVoidedNode{IRAssignNode{reg, child}}
}

func transformReturnStmt(s scope.Scope, stmt *parser.ReturnStmt) IRVoidNode {
	if stmt.Expr != nil {
		return IRReturnNode{transformExpr(s, stmt.Expr)}
	}

	return IRReturnNode{nil}
}

func transformExprStmt(s scope.Scope, stmt *parser.ExprStmt) IRVoidNode {
	return IRVoidedNode{transformExpr(s, stmt.Expr)}
}

func transformExpr(s scope.Scope, expr parser.Expr) IRTypedNode {
	switch expr := expr.(type) {
	case *parser.FunctionExpr:
		return transformFunctionExpr(s, expr)
	case *parser.DispatchExpr:
		return transformDispatchExpr(s, expr)
	case *parser.AssignExpr:
		return transformAssignExpr(s, expr)
	case *parser.BinaryExpr:
		return transformBinaryExpr(s, expr)
	case *parser.SelfExpr:
		return transformSelfExpr(s, expr)
	case *parser.IdentExpr:
		return transformIdentExpr(s, expr)
	case *parser.NumberExpr:
		return transformNumberExpr(s, expr)
	case *parser.StringExpr:
		return transformStringExpr(s, expr)
	case *parser.BooleanExpr:
		return transformBooleanExpr(s, expr)
	default:
		panic(fmt.Sprintf("cannot transform %T", expr))
	}
}

func transformFunctionExpr(s scope.Scope, expr *parser.FunctionExpr) IRTypedNode {
	local := expr.Scope
	var params []*vm.RegisterTemplate
	for _, param := range expr.Params {
		name := param.Name.Name
		reg := local.GetLocalVariableRegister(name)
		params = append(params, reg)
	}

	block := transformStmts(local, expr.Block.Stmts)
	block = append(block, IRReturnNode{})
	return IRFunctionNode{local, params, local.GetSelfReference().Ret, block}
}

func transformDispatchExpr(s scope.Scope, expr *parser.DispatchExpr) IRTypedNode {
	callee := transformExpr(s, expr.Callee)

	var args []IRTypedNode
	for _, expr := range expr.Args {
		arg := transformExpr(s, expr)
		args = append(args, arg)
	}

	return IRDispatchNode{callee, args}
}

func transformAssignExpr(s scope.Scope, expr *parser.AssignExpr) IRTypedNode {
	name := expr.Left.Name
	child := transformExpr(s, expr.Right)
	reg := s.GetVariableRegister(name)
	return IRAssignNode{reg, child}
}

func transformBinaryExpr(s scope.Scope, expr *parser.BinaryExpr) IRTypedNode {
	oper := expr.Oper
	left := transformExpr(s, expr.Left)
	right := transformExpr(s, expr.Right)
	return IRBinaryNode{oper, left, right}
}

func transformSelfExpr(s scope.Scope, expr *parser.SelfExpr) IRTypedNode {
	return IRSelfReferenceNode{s.GetSelfReference().Ret}
}

func transformIdentExpr(s scope.Scope, expr *parser.IdentExpr) IRTypedNode {
	name := expr.Name
	reg := s.GetVariableRegister(name)
	return IRReferenceNode{reg}
}

func transformNumberExpr(s scope.Scope, expr *parser.NumberExpr) IRTypedNode {
	return IRIntegerLiteralNode{int64(expr.Val)}
}

func transformStringExpr(s scope.Scope, expr *parser.StringExpr) IRTypedNode {
	return IRStringLiteralNode{expr.Val}
}

func transformBooleanExpr(s scope.Scope, expr *parser.BooleanExpr) IRTypedNode {
	return IRBooleanLitearlNode{expr.Val}
}
