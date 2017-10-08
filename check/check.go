package check

import (
	"fmt"
	"plaid/parser"
	"plaid/types"
)

// A collection of types native to the execution environment
var (
	BuiltinInt types.Type = types.TypeIdent{Name: "Int"}
	BuiltinStr            = types.TypeIdent{Name: "Str"}
)

// Check takes an existing abstract syntax tree and performs type checks and
// other correctness checks. It returns a list of any errors that were
// discovered inside the AST
func Check(prog parser.Program) *Scope {
	global := makeScope(nil, nil)
	checkProgram(global, prog)
	return global
}

func checkProgram(scope *Scope, prog parser.Program) {
	for _, stmt := range prog.Stmts {
		checkStmt(scope, stmt)
	}
}

func checkStmt(scope *Scope, stmt parser.Stmt) {
	switch stmt := stmt.(type) {
	case parser.DeclarationStmt:
		checkDeclarationStmt(scope, stmt)
		break
	case parser.PrintStmt:
		checkPrintStmt(scope, stmt)
		break
	case parser.ReturnStmt:
		checkReturnStmt(scope, stmt)
		break
	case parser.ExprStmt:
		checkExprAllowVoid(scope, stmt.Expr)
		break
	}

	if scope.hasBodyQueue() {
		sig, expr := scope.dequeueBody()
		checkFunctionBody(scope, sig, expr)
	}
}

func checkStmtBlock(scope *Scope, block parser.StmtBlock) {
	for _, stmt := range block.Stmts {
		checkStmt(scope, stmt)
	}
}

func checkDeclarationStmt(scope *Scope, stmt parser.DeclarationStmt) {
	name := stmt.Name.Name
	typ := checkExpr(scope, stmt.Expr)
	scope.registerLocalVariable(name, typ)
}

func checkPrintStmt(scope *Scope, stmt parser.PrintStmt) {
	if stmt.Expr == nil {
		scope.addError(fmt.Errorf("expected an expression to print"))
		return
	}

	checkExpr(scope, stmt.Expr)
}

func checkReturnStmt(scope *Scope, stmt parser.ReturnStmt) {
	var ret types.Type = types.TypeVoid{}
	if stmt.Expr != nil {
		ret = checkExpr(scope, stmt.Expr)
	}

	if scope.pendingReturn == nil {
		scope.addError(fmt.Errorf("return statements must be inside a function"))
		return
	}

	if scope.pendingReturn.Equals(ret) || ret.IsError() {
		return
	}

	if scope.pendingReturn.Equals(types.TypeVoid{}) {
		scope.addError(fmt.Errorf("expected to return nothing, got '%s'", ret))
		return
	}

	if ret.Equals(types.TypeVoid{}) {
		scope.addError(fmt.Errorf("expected a return type of '%s', got nothing", scope.pendingReturn))
	}

	scope.addError(fmt.Errorf("expected to return '%s', got '%s'", scope.pendingReturn, ret))
}

func checkExprAllowVoid(scope *Scope, expr parser.Expr) types.Type {
	var typ types.Type = types.TypeError{}
	switch expr := expr.(type) {
	case parser.FunctionExpr:
		typ = checkFunctionExpr(scope, expr)
	case parser.DispatchExpr:
		typ = checkDispatchExpr(scope, expr)
	case parser.AssignExpr:
		typ = checkAssignExpr(scope, expr)
	case parser.BinaryExpr:
		typ = checkBinaryExpr(scope, expr)
	case parser.IdentExpr:
		typ = checkIdentExpr(scope, expr)
	case parser.NumberExpr:
		typ = checkNumberExpr(scope, expr)
	case parser.StringExpr:
		typ = checkStringExpr(scope, expr)
	default:
		scope.addError(fmt.Errorf("unknown expression type"))
	}

	return typ
}

func checkExpr(scope *Scope, expr parser.Expr) types.Type {
	typ := checkExprAllowVoid(scope, expr)

	if typ.Equals(types.TypeVoid{}) {
		scope.addError(fmt.Errorf("cannot use void types in an expression"))
		return types.TypeError{}
	}

	return typ
}

func checkFunctionExpr(scope *Scope, expr parser.FunctionExpr) types.Type {
	ret := types.ConvertTypeNote(expr.Ret)
	params := []types.Type{}
	for _, param := range expr.Params {
		params = append(params, types.ConvertTypeNote(param.Note))
	}
	tuple := types.TypeTuple{Children: params}

	sig := types.TypeFunction{Params: tuple, Ret: ret}
	scope.enqueueBody(ret, expr)
	return sig
}

func checkFunctionBody(scope *Scope, ret types.Type, expr parser.FunctionExpr) {
	pushed := makeScope(scope, ret)

	for _, param := range expr.Params {
		paramName := param.Name.Name
		paramType := types.ConvertTypeNote(param.Note)
		pushed.registerLocalVariable(paramName, paramType)
	}

	checkStmtBlock(pushed, expr.Block)
}

func checkDispatchExpr(scope *Scope, expr parser.DispatchExpr) types.Type {
	// Resolve arguments to types
	argTypes := []types.Type{}
	for _, argExpr := range expr.Args {
		argTypes = append(argTypes, checkExpr(scope, argExpr))
	}

	// Resolve callee to type
	calleeType := checkExpr(scope, expr.Callee)
	calleeFunc, ok := calleeType.(types.TypeFunction)
	if ok == false {
		if calleeType.IsError() == false {
			scope.addError(fmt.Errorf("cannot call function on type '%s'", calleeType))
		}

		return types.TypeError{}
	}

	// Resolve return type
	retType := calleeFunc.Ret

	// Check that the given argument types match the expected parameter types
	totalArgs := len(argTypes)
	totalParams := len(calleeFunc.Params.Children)
	if totalArgs == totalParams {
		for i := 0; i < totalArgs; i++ {
			argType := argTypes[i]
			paramType := calleeFunc.Params.Children[i]

			if argType.Equals(paramType) == false {
				scope.addError(fmt.Errorf("expected '%s', got '%s'", paramType, argType))
				retType = types.TypeError{}
			}
		}
	} else {
		scope.addError(fmt.Errorf("expected %d arguments, got %d", totalParams, totalArgs))
		retType = types.TypeError{}
	}

	return retType
}

func checkAssignExpr(scope *Scope, expr parser.AssignExpr) types.Type {
	name := expr.Left.Name
	leftType := scope.getVariable(name)
	rightType := checkExpr(scope, expr.Right)

	if leftType == nil {
		scope.addError(fmt.Errorf("'%s' cannot be assigned before it is declared", name))
		return types.TypeError{}
	}

	if leftType.IsError() || rightType.IsError() {
		return types.TypeError{}
	}

	if leftType.Equals(rightType) == false {
		scope.addError(fmt.Errorf("'%s' cannot be assigned type '%s'", leftType, rightType))
		return types.TypeError{}
	}

	return leftType
}

func checkBinaryExpr(scope *Scope, expr parser.BinaryExpr) types.Type {
	switch expr.Oper {
	case "+":
		fallthrough
	case "-":
		fallthrough
	case "*":
		return expectBinaryTypes(scope, expr.Left, BuiltinInt, expr.Right, BuiltinInt, BuiltinInt)
	default:
		scope.addError(fmt.Errorf("unknown infix operator '%s'", expr.Oper))
		return types.TypeError{}
	}
}

func expectBinaryTypes(scope *Scope, left parser.Expr, expLeftType types.Type, right parser.Expr, expRightType types.Type, retType types.Type) types.Type {
	leftType := checkExpr(scope, left)
	rightType := checkExpr(scope, right)

	if leftType.IsError() || rightType.IsError() {
		return types.TypeError{}
	}

	typ := retType

	if leftType.Equals(expLeftType) == false {
		scope.addError(fmt.Errorf("left side must have type %s, got %s", expLeftType, leftType))
		typ = types.TypeError{}
	}

	if rightType.Equals(expRightType) == false {
		scope.addError(fmt.Errorf("right side must have type %s, got %s", expRightType, rightType))
		typ = types.TypeError{}
	}

	return typ
}

func checkIdentExpr(scope *Scope, expr parser.IdentExpr) types.Type {
	if scope.existingVariable(expr.Name) {
		return scope.getVariable(expr.Name)
	}

	scope.addError(fmt.Errorf("variable '%s' was used before it was declared", expr.Name))
	return types.TypeError{}
}

func checkNumberExpr(scope *Scope, expr parser.NumberExpr) types.Type {
	return BuiltinInt
}

func checkStringExpr(scope *Scope, expr parser.StringExpr) types.Type {
	return BuiltinStr
}
