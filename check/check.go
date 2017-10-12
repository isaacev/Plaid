package check

import (
	"fmt"
	"plaid/parser"
	"plaid/scope"
	"plaid/types"
	"plaid/vm"
)

type binopsLUT map[string]map[types.Type]map[types.Type]types.Type
type doubleLUT map[types.Type]map[types.Type]types.Type
type singleLUT map[types.Type]types.Type

var defaultBinopsLUT = binopsLUT{
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
		types.Str: singleLUT{types.Int: types.TypeOptional{Child: types.Str}},
	},
}

// Check takes an existing abstract syntax tree and performs type checks and
// other correctness checks. It returns a list of any errors that were
// discovered inside the AST
func Check(prog *parser.Program, libraries ...vm.Library) *scope.GlobalScope {
	global := scope.MakeGlobalScope()

	for _, library := range libraries {
		for name, builtin := range library {
			global.NewVariable(name, builtin.Type)
		}
	}

	checkProgram(global, prog)
	return global
}

func checkProgram(s scope.Scope, prog *parser.Program) {
	for _, stmt := range prog.Stmts {
		checkStmt(s, stmt)
	}
}

func checkStmt(s scope.Scope, stmt parser.Stmt) {
	switch stmt := stmt.(type) {
	case *parser.IfStmt:
		checkIfStmt(s, stmt)
		break
	case *parser.DeclarationStmt:
		checkDeclarationStmt(s, stmt)
		break
	case *parser.ReturnStmt:
		checkReturnStmt(s, stmt)
		break
	case *parser.ExprStmt:
		checkExprAllowVoid(s, stmt.Expr)
		break
	}
}

func checkStmtBlock(s scope.Scope, block *parser.StmtBlock) {
	for _, stmt := range block.Stmts {
		checkStmt(s, stmt)
	}
}

func checkIfStmt(s scope.Scope, stmt *parser.IfStmt) {
	typ := checkExpr(s, stmt.Cond)
	if typ.Equals(types.Bool) == false {
		s.NewError(fmt.Errorf("condition must resolve to a boolean"))
	}

	checkStmtBlock(s, stmt.Clause)
}

func checkDeclarationStmt(s scope.Scope, stmt *parser.DeclarationStmt) {
	name := stmt.Name.Name
	typ := checkExpr(s, stmt.Expr)
	s.NewVariable(name, typ)
}

func checkReturnStmt(s scope.Scope, stmt *parser.ReturnStmt) {
	var ret types.Type = types.TypeVoid{}
	if stmt.Expr != nil {
		ret = checkExpr(s, stmt.Expr)
	}

	if s.HasSelfReference() == false {
		s.NewError(fmt.Errorf("return statements must be inside a function"))
		return
	}

	if s.GetSelfReference().Ret.Equals(ret) || ret.IsError() {
		return
	}

	if s.GetSelfReference().Ret.Equals(types.TypeVoid{}) {
		s.NewError(fmt.Errorf("expected to return nothing, got '%s'", ret))
		return
	}

	if ret.Equals(types.TypeVoid{}) {
		s.NewError(fmt.Errorf("expected a return type of '%s', got nothing", s.GetSelfReference().Ret))
	}

	s.NewError(fmt.Errorf("expected to return '%s', got '%s'", s.GetSelfReference().Ret, ret))
}

func checkExprAllowVoid(s scope.Scope, expr parser.Expr) types.Type {
	var typ types.Type = types.TypeError{}
	switch expr := expr.(type) {
	case *parser.FunctionExpr:
		typ = checkFunctionExpr(s, expr)
	case *parser.DispatchExpr:
		typ = checkDispatchExpr(s, expr)
	case *parser.AssignExpr:
		typ = checkAssignExpr(s, expr)
	case *parser.BinaryExpr:
		typ = checkBinaryExpr(s, expr, defaultBinopsLUT)
	case *parser.ListExpr:
		typ = checkListExpr(s, expr)
	case *parser.SubscriptExpr:
		typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	case *parser.SelfExpr:
		typ = checkSelfExpr(s, expr)
	case *parser.IdentExpr:
		typ = checkIdentExpr(s, expr)
	case *parser.NumberExpr:
		typ = checkNumberExpr(s, expr)
	case *parser.StringExpr:
		typ = checkStringExpr(s, expr)
	case *parser.BooleanExpr:
		typ = checkBooleanExpr(s, expr)
	default:
		s.NewError(fmt.Errorf("unknown expression type"))
	}

	return typ
}

func checkExpr(s scope.Scope, expr parser.Expr) types.Type {
	typ := checkExprAllowVoid(s, expr)

	if typ.Equals(types.TypeVoid{}) {
		s.NewError(fmt.Errorf("cannot use void types in an expression"))
		return types.TypeError{}
	}

	return typ
}

func checkFunctionExpr(s scope.Scope, expr *parser.FunctionExpr) types.Type {
	ret := ConvertTypeNote(expr.Ret)
	params := []types.Type{}
	for _, param := range expr.Params {
		params = append(params, ConvertTypeNote(param.Note))
	}
	tuple := types.TypeTuple{Children: params}
	self := types.TypeFunction{Params: tuple, Ret: ret}

	childScope := scope.MakeLocalScope(s, self)

	for _, param := range expr.Params {
		paramName := param.Name.Name
		paramType := ConvertTypeNote(param.Note)
		childScope.NewVariable(paramName, paramType)
	}

	checkStmtBlock(childScope, expr.Block)
	expr.Scope = childScope
	return self
}

func checkDispatchExpr(s scope.Scope, expr *parser.DispatchExpr) types.Type {
	// Resolve arguments to types
	argTypes := []types.Type{}
	for _, argExpr := range expr.Args {
		argTypes = append(argTypes, checkExpr(s, argExpr))
	}

	// Resolve callee to type
	calleeType := checkExpr(s, expr.Callee)
	calleeFunc, ok := calleeType.(types.TypeFunction)
	if ok == false {
		if calleeType.IsError() == false {
			s.NewError(fmt.Errorf("cannot call function on type '%s'", calleeType))
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
				s.NewError(fmt.Errorf("expected '%s', got '%s'", paramType, argType))
				retType = types.TypeError{}
			}
		}
	} else {
		s.NewError(fmt.Errorf("expected %d arguments, got %d", totalParams, totalArgs))
		retType = types.TypeError{}
	}

	return retType
}

func checkAssignExpr(s scope.Scope, expr *parser.AssignExpr) types.Type {
	name := expr.Left.Name
	leftType := s.GetVariableType(name)
	rightType := checkExpr(s, expr.Right)

	if leftType == nil {
		s.NewError(fmt.Errorf("'%s' cannot be assigned before it is declared", name))
		return types.TypeError{}
	}

	if leftType.IsError() || rightType.IsError() {
		return types.TypeError{}
	}

	if leftType.Equals(rightType) == false {
		s.NewError(fmt.Errorf("'%s' cannot be assigned type '%s'", leftType, rightType))
		return types.TypeError{}
	}

	return leftType
}

func checkBinaryExpr(s scope.Scope, expr *parser.BinaryExpr, lut binopsLUT) types.Type {
	leftType := checkExpr(s, expr.Left)
	rightType := checkExpr(s, expr.Right)

	if leftType.IsError() || rightType.IsError() {
		return types.TypeError{}
	}

	if operLUT, ok := lut[expr.Oper]; ok {
		if leftLUT, ok := operLUT[leftType]; ok {
			if retType, ok := leftLUT[rightType]; ok {
				return retType
			}
		}

		s.NewError(fmt.Errorf("operator '%s' does not support %s and %s", expr.Oper, leftType, rightType))
		return types.TypeError{}
	}

	s.NewError(fmt.Errorf("unknown infix operator '%s'", expr.Oper))
	return types.TypeError{}
}

func checkListExpr(s scope.Scope, expr *parser.ListExpr) types.Type {
	var elemTypes []types.Type
	for _, elem := range expr.Elements {
		elemTypes = append(elemTypes, checkExpr(s, elem))
	}

	if len(elemTypes) == 0 {
		s.NewError(fmt.Errorf("cannot determine type from empty list"))
		return types.TypeError{}
	}

	for _, typ := range elemTypes {
		if typ.IsError() {
			return types.TypeError{}
		}
	}

	var listType types.Type
	for _, typ := range elemTypes {
		if listType == nil {
			listType = typ
			continue
		}

		if listType.Equals(typ) == false {
			s.NewError(fmt.Errorf("element type %s is not compatible with type %s", typ, listType))
			return types.TypeError{}
		}
	}

	return types.TypeList{Child: listType}
}

func checkSubscriptExpr(s scope.Scope, expr *parser.SubscriptExpr, lut binopsLUT) types.Type {
	listType := checkExpr(s, expr.ListLike)
	indexType := checkExpr(s, expr.Index)

	if listType.IsError() || indexType.IsError() {
		return types.TypeError{}
	}

	if listType, ok := listType.(types.TypeList); ok {
		return types.TypeOptional{Child: listType.Child}
	}

	if subscriptLUT, ok := lut["["]; ok {
		if listLUT, ok := subscriptLUT[listType]; ok {
			if retType, ok := listLUT[indexType]; ok {
				return retType
			}
		}

		s.NewError(fmt.Errorf("subscript operator does not support %s[%s]", listType, indexType))
		return types.TypeError{}
	}

	s.NewError(fmt.Errorf("unknown infix operator '['"))
	return types.TypeError{}
}

func checkSelfExpr(s scope.Scope, expr *parser.SelfExpr) types.Type {
	if s.HasSelfReference() == false {
		s.NewError(fmt.Errorf("self references must be inside a function"))
		return types.TypeError{}
	}

	return s.GetSelfReference()
}

func checkIdentExpr(s scope.Scope, expr *parser.IdentExpr) types.Type {
	if s.HasVariable(expr.Name) {
		return s.GetVariableType(expr.Name)
	}

	s.NewError(fmt.Errorf("variable '%s' was used before it was declared", expr.Name))
	return types.TypeError{}
}

func checkNumberExpr(s scope.Scope, expr *parser.NumberExpr) types.Type {
	return types.Int
}

func checkStringExpr(s scope.Scope, expr *parser.StringExpr) types.Type {
	return types.Str
}

func checkBooleanExpr(s scope.Scope, expr *parser.BooleanExpr) types.Type {
	return types.Bool
}
