package codegen

import (
	"fmt"
	"plaid/parser"
	"plaid/types"
	"plaid/vm"
)

// Transform converts an AST to the intermediate representation (IR) which is a
// precursor to the compiled bytecode
func Transform(prog parser.Program, libraries ...vm.Library) IR {
	scope := makeLexicalScope(nil)

	for _, library := range libraries {
		for name, builtin := range library {
			scope.addGlobalVariable(name, builtin)
		}
	}

	nodes := transformStmts(scope, prog.Stmts)
	return IR{scope, nodes}
}

func transformStmts(scope *LexicalScope, stmts []parser.Stmt) []IRVoidNode {
	var nodes []IRVoidNode
	for _, stmt := range stmts {
		node := transformStmt(scope, stmt)
		nodes = append(nodes, node)
	}
	return nodes
}

func transformStmt(scope *LexicalScope, stmt parser.Stmt) IRVoidNode {
	switch stmt := stmt.(type) {
	case parser.IfStmt:
		return transformIfStmt(scope, stmt)
	case parser.DeclarationStmt:
		return transformDeclarationStmt(scope, stmt)
	case parser.ReturnStmt:
		return transformReturnStmt(scope, stmt)
	case parser.ExprStmt:
		return transformExprStmt(scope, stmt)
	default:
		panic(fmt.Sprintf("cannot transform %T", stmt))
	}
}

func transformIfStmt(scope *LexicalScope, stmt parser.IfStmt) IRVoidNode {
	cond := transformExpr(scope, stmt.Cond)
	clause := transformStmts(scope, stmt.Clause.Stmts)
	return IRCondNode{cond, clause}
}

func transformDeclarationStmt(scope *LexicalScope, stmt parser.DeclarationStmt) IRVoidNode {
	name := stmt.Name.Name
	child := transformExpr(scope, stmt.Expr)
	record := scope.addLocalVariable(name, child.Type())
	return IRVoidedNode{IRAssignNode{record, child}}
}

func transformReturnStmt(scope *LexicalScope, stmt parser.ReturnStmt) IRVoidNode {
	if stmt.Expr != nil {
		return IRReturnNode{transformExpr(scope, stmt.Expr)}
	}

	return IRReturnNode{nil}
}

func transformExprStmt(scope *LexicalScope, stmt parser.ExprStmt) IRVoidNode {
	return IRVoidedNode{transformExpr(scope, stmt.Expr)}
}

func transformExpr(scope *LexicalScope, expr parser.Expr) IRTypedNode {
	switch expr := expr.(type) {
	case parser.FunctionExpr:
		return transformFunctionExpr(scope, expr)
	case parser.DispatchExpr:
		return transformDispatchExpr(scope, expr)
	case parser.AssignExpr:
		return transformAssignExpr(scope, expr)
	case parser.BinaryExpr:
		return transformBinaryExpr(scope, expr)
	case parser.SelfExpr:
		return transformSelfExpr(scope, expr)
	case parser.IdentExpr:
		return transformIdentExpr(scope, expr)
	case parser.NumberExpr:
		return transformNumberExpr(scope, expr)
	case parser.StringExpr:
		return transformStringExpr(scope, expr)
	case parser.BooleanExpr:
		return transformBooleanExpr(scope, expr)
	default:
		panic(fmt.Sprintf("cannot transform %T", expr))
	}
}

func transformFunctionExpr(scope *LexicalScope, expr parser.FunctionExpr) IRTypedNode {
	local := makeLexicalScope(scope)

	var params []*VarRecord
	for _, param := range expr.Params {
		name := param.Name.Name
		typ := types.ConvertTypeNote(param.Note)
		record := local.addLocalVariable(name, typ)
		params = append(params, record)
	}

	ret := types.ConvertTypeNote(expr.Ret)
	local.SelfRet = ret
	block := transformStmts(local, expr.Block.Stmts)
	block = append(block, IRReturnNode{})
	return IRFunctionNode{local, params, ret, block}
}

func transformDispatchExpr(scope *LexicalScope, expr parser.DispatchExpr) IRTypedNode {
	callee := transformExpr(scope, expr.Callee)

	var args []IRTypedNode
	for _, expr := range expr.Args {
		arg := transformExpr(scope, expr)
		args = append(args, arg)
	}

	return IRDispatchNode{callee, args}
}

func transformAssignExpr(scope *LexicalScope, expr parser.AssignExpr) IRTypedNode {
	name := expr.Left.Name
	child := transformExpr(scope, expr.Right)
	record := scope.getVariable(name)
	return IRAssignNode{record, child}
}

func transformBinaryExpr(scope *LexicalScope, expr parser.BinaryExpr) IRTypedNode {
	oper := expr.Oper
	left := transformExpr(scope, expr.Left)
	right := transformExpr(scope, expr.Right)
	return IRBinaryNode{oper, left, right}
}

func transformSelfExpr(scope *LexicalScope, expr parser.SelfExpr) IRTypedNode {
	return IRSelfReferenceNode{scope.SelfRet}
}

func transformIdentExpr(scope *LexicalScope, expr parser.IdentExpr) IRTypedNode {
	name := expr.Name
	if scope.hasLocalVariable(name) {
		record := scope.getVariable(name)
		return IRReferenceNode{record}
	} else if scope.hasGlobalVariable(name) {
		record := scope.getGlobalVariable(name)
		return IRBuiltinReferenceNode{record}
	} else {
		record := scope.getVariable(name)
		return IRReferenceNode{record}
	}
}

func transformNumberExpr(scope *LexicalScope, expr parser.NumberExpr) IRTypedNode {
	return IRIntegerLiteralNode{int64(expr.Val)}
}

func transformStringExpr(scope *LexicalScope, expr parser.StringExpr) IRTypedNode {
	return IRStringLiteralNode{expr.Val}
}

func transformBooleanExpr(scope *LexicalScope, expr parser.BooleanExpr) IRTypedNode {
	return IRBooleanLitearlNode{expr.Val}
}
