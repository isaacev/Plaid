package check

import (
	"fmt"
	"plaid/parser"
	"plaid/types"
	"plaid/vm"
)

type binopsLUT map[string]map[types.Type]map[types.Type]types.Type
type doubleLUT map[types.Type]map[types.Type]types.Type
type singleLUT map[types.Type]types.Type

var binops = binopsLUT{
	"+": doubleLUT{
		types.Int: singleLUT{types.Int: types.Int},
		types.Str: singleLUT{types.Str: types.Str},
	},
	"-": doubleLUT{
		types.Int: singleLUT{types.Int: types.Int},
	},
	"*": doubleLUT{
		types.Int: singleLUT{types.Int: types.Int},
	},
	"<": doubleLUT{
		types.Int: singleLUT{types.Int: types.Bool},
	},
	"<=": doubleLUT{
		types.Int: singleLUT{types.Int: types.Bool},
	},
	">": doubleLUT{
		types.Int: singleLUT{types.Int: types.Bool},
	},
	">=": doubleLUT{
		types.Int: singleLUT{types.Int: types.Bool},
	},
	"[": doubleLUT{
		types.Str: singleLUT{types.Int: types.Str},
	},
}

// Check takes an existing abstract syntax tree and performs type checks and
// other correctness checks. It returns a list of any errors that were
// discovered inside the AST
func Check(prog parser.Program, libraries ...vm.Library) *Scope {
	global := makeScope(nil, nil)

	for _, library := range libraries {
		for name, builtin := range library {
			global.registerLocalVariable(name, builtin.Type)
		}
	}

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
	case parser.IfStmt:
		checkIfStmt(scope, stmt)
		break
	case parser.DeclarationStmt:
		checkDeclarationStmt(scope, stmt)
		break
	case parser.ReturnStmt:
		checkReturnStmt(scope, stmt)
		break
	case parser.ExprStmt:
		checkExprAllowVoid(scope, stmt.Expr)
		break
	}
}

func checkStmtBlock(scope *Scope, block parser.StmtBlock) {
	for _, stmt := range block.Stmts {
		checkStmt(scope, stmt)
	}
}

func checkIfStmt(scope *Scope, stmt parser.IfStmt) {
	typ := checkExpr(scope, stmt.Cond)
	if typ.Equals(types.Bool) == false {
		scope.addError(fmt.Errorf("condition must resolve to a boolean"))
	}

	checkStmtBlock(scope, stmt.Clause)
}

func checkDeclarationStmt(scope *Scope, stmt parser.DeclarationStmt) {
	name := stmt.Name.Name
	typ := checkExpr(scope, stmt.Expr)
	scope.registerLocalVariable(name, typ)
}

func checkReturnStmt(scope *Scope, stmt parser.ReturnStmt) {
	var ret types.Type = types.TypeVoid{}
	if stmt.Expr != nil {
		ret = checkExpr(scope, stmt.Expr)
	}

	if scope.self == nil {
		scope.addError(fmt.Errorf("return statements must be inside a function"))
		return
	}

	if scope.self.Equals(ret) || ret.IsError() {
		return
	}

	if scope.self.Ret.Equals(types.TypeVoid{}) {
		scope.addError(fmt.Errorf("expected to return nothing, got '%s'", ret))
		return
	}

	if ret.Equals(types.TypeVoid{}) {
		scope.addError(fmt.Errorf("expected a return type of '%s', got nothing", scope.self.Ret))
	}

	scope.addError(fmt.Errorf("expected to return '%s', got '%s'", scope.self.Ret, ret))
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
	case parser.SubscriptExpr:
		typ = checkSubscriptExpr(scope, expr)
	case parser.SelfExpr:
		typ = checkSelfExpr(scope, expr)
	case parser.IdentExpr:
		typ = checkIdentExpr(scope, expr)
	case parser.NumberExpr:
		typ = checkNumberExpr(scope, expr)
	case parser.StringExpr:
		typ = checkStringExpr(scope, expr)
	case parser.BooleanExpr:
		typ = checkBooleanExpr(scope, expr)
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
	self := types.TypeFunction{Params: tuple, Ret: ret}

	childScope := makeScope(scope, &self)

	for _, param := range expr.Params {
		paramName := param.Name.Name
		paramType := types.ConvertTypeNote(param.Note)
		childScope.registerLocalVariable(paramName, paramType)
	}

	checkStmtBlock(childScope, expr.Block)
	return self
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

			if argType.IsError() {
				retType = types.TypeError{}
			} else if argType.Equals(paramType) == false {
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
	leftType := checkExpr(scope, expr.Left)
	rightType := checkExpr(scope, expr.Right)

	if leftType.IsError() || rightType.IsError() {
		return types.TypeError{}
	}

	if operLUT, ok := binops[expr.Oper]; ok {
		if leftLUT, ok := operLUT[leftType]; ok {
			if retType, ok := leftLUT[rightType]; ok {
				return retType
			}
		}

		scope.addError(fmt.Errorf("operator '%s' does not support %s and %s", expr.Oper, leftType, rightType))
		return types.TypeError{}
	}

	scope.addError(fmt.Errorf("unknown infix operator '%s'", expr.Oper))
	return types.TypeError{}
}

func checkSubscriptExpr(scope *Scope, expr parser.SubscriptExpr) types.Type {
	listType := checkExpr(scope, expr.ListLike)
	indexType := checkExpr(scope, expr.Index)

	if listType.IsError() || indexType.IsError() {
		return types.TypeError{}
	}

	if subscriptLUT, ok := binops["["]; ok {
		if listLUT, ok := subscriptLUT[listType]; ok {
			if retType, ok := listLUT[indexType]; ok {
				return retType
			}
		}

		scope.addError(fmt.Errorf("subscript operator does not support %s[%s]", listType, indexType))
		return types.TypeError{}
	}

	scope.addError(fmt.Errorf("unknown infix operator '['"))
	return types.TypeError{}
}

func checkSelfExpr(scope *Scope, expr parser.SelfExpr) types.Type {
	if scope.self == nil {
		scope.addError(fmt.Errorf("self references must be inside a function"))
		return types.TypeError{}
	}

	return *scope.self
}

func checkIdentExpr(scope *Scope, expr parser.IdentExpr) types.Type {
	if scope.existingVariable(expr.Name) {
		return scope.getVariable(expr.Name)
	}

	scope.addError(fmt.Errorf("variable '%s' was used before it was declared", expr.Name))
	return types.TypeError{}
}

func checkNumberExpr(scope *Scope, expr parser.NumberExpr) types.Type {
	return types.Int
}

func checkStringExpr(scope *Scope, expr parser.StringExpr) types.Type {
	return types.Str
}

func checkBooleanExpr(scope *Scope, expr parser.BooleanExpr) types.Type {
	return types.Bool
}
