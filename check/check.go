package check

import (
	"fmt"
	"plaid/lexer"
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
		types.Str: singleLUT{types.Int: types.Optional{Child: types.Str}},
	},
}

// Check takes an existing abstract syntax tree and performs type checks and
// other correctness checks. It returns a list of any errors that were
// discovered inside the AST
func Check(prog *parser.Program, modules ...*vm.Module) *scope.GlobalScope {
	global := scope.MakeGlobalScope()

	for _, module := range modules {
		global.Import(module)
	}

	checkProgram(global, prog)
	prog.Scope = global
	return global
}

func checkProgram(s scope.Scope, prog *parser.Program) {
	for _, stmt := range prog.Stmts {
		checkStmt(s, stmt)
	}
}

func checkStmt(s scope.Scope, stmt parser.Stmt) {
	switch stmt := stmt.(type) {
	case *parser.PubStmt:
		checkPubStmt(s, stmt)
		break
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

func checkPubStmt(s scope.Scope, stmt *parser.PubStmt) {
	checkStmt(s, stmt.Stmt)

	var g *scope.GlobalScope
	var ok bool
	if g, ok = s.(*scope.GlobalScope); ok == false {
		addTypeError(s, stmt.Start(), "pub statement must be a top-level statement")
		return
	}

	name := stmt.Stmt.Name.Name
	typ := g.GetVariableType(name)
	g.Export(name, typ)
}

func checkIfStmt(s scope.Scope, stmt *parser.IfStmt) {
	typ := checkExpr(s, stmt.Cond)
	if types.Bool.Equals(typ) == false {
		addTypeError(s, stmt.Cond.Start(), "condition must resolve to a boolean")
	}

	checkStmtBlock(s, stmt.Clause)
}

func checkDeclarationStmt(s scope.Scope, stmt *parser.DeclarationStmt) {
	name := stmt.Name.Name
	typ := checkExpr(s, stmt.Expr)
	s.NewVariable(name, typ)
}

func checkReturnStmt(s scope.Scope, stmt *parser.ReturnStmt) {
	var ret types.Type = types.Void{}
	if stmt.Expr != nil {
		ret = checkExpr(s, stmt.Expr)
	}

	if s.HasSelfReference() == false {
		addTypeError(s, stmt.Start(), "return statements must be inside a function")
		return
	}

	if s.GetSelfReference().Ret.Equals(ret) || ret.IsError() {
		return
	}

	if s.GetSelfReference().Ret.Equals(types.Void{}) {
		msg := fmt.Sprintf("expected to return nothing, got '%s'", ret)
		addTypeError(s, stmt.Expr.Start(), msg)
		return
	}

	if (types.Void{}).Equals(ret) {
		msg := fmt.Sprintf("expected a return type of '%s', got nothing", s.GetSelfReference().Ret)
		addTypeError(s, stmt.Start(), msg)
		return
	}

	msg := fmt.Sprintf("expected to return '%s', got '%s'", s.GetSelfReference().Ret, ret)
	addTypeError(s, stmt.Expr.Start(), msg)
}

func checkExprAllowVoid(s scope.Scope, expr parser.Expr) types.Type {
	var typ types.Type = types.Error{}
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
		addTypeError(s, expr.Start(), "unknown expression type")
	}

	return typ
}

func checkExpr(s scope.Scope, expr parser.Expr) types.Type {
	typ := checkExprAllowVoid(s, expr)

	if (types.Void{}).Equals(typ) {
		addTypeError(s, expr.Start(), "cannot use void types in an expression")
		return types.Error{}
	}

	return typ
}

func checkFunctionExpr(s scope.Scope, expr *parser.FunctionExpr) types.Type {
	ret := ConvertTypeNote(expr.Ret)
	params := []types.Type{}
	for _, param := range expr.Params {
		params = append(params, ConvertTypeNote(param.Note))
	}
	tuple := types.Tuple{Children: params}
	self := types.Function{Params: tuple, Ret: ret}

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
	calleeFunc, ok := calleeType.(types.Function)
	if ok == false {
		if calleeType.IsError() == false {
			msg := fmt.Sprintf("cannot call function on type '%s'", calleeType)
			addTypeError(s, expr.Start(), msg)
		}

		return types.Error{}
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
				retType = types.Error{}
			} else if paramType.Equals(argType) == false {
				msg := fmt.Sprintf("expected '%s', got '%s'", paramType, argType)
				addTypeError(s, expr.Args[i].Start(), msg)
				retType = types.Error{}
			}
		}
	} else {
		msg := fmt.Sprintf("expected %d arguments, got %d", totalParams, totalArgs)
		addTypeError(s, expr.Start(), msg)
		retType = types.Error{}
	}

	return retType
}

func checkAssignExpr(s scope.Scope, expr *parser.AssignExpr) types.Type {
	name := expr.Left.Name
	leftType := s.GetVariableType(name)
	rightType := checkExpr(s, expr.Right)

	if leftType == nil {
		msg := fmt.Sprintf("'%s' cannot be assigned before it is declared", name)
		addTypeError(s, expr.Start(), msg)
		return types.Error{}
	}

	if leftType.IsError() || rightType.IsError() {
		return types.Error{}
	}

	if leftType.Equals(rightType) == false {
		msg := fmt.Sprintf("'%s' cannot be assigned type '%s'", leftType, rightType)
		addTypeError(s, expr.Right.Start(), msg)
		return types.Error{}
	}

	return leftType
}

func checkBinaryExpr(s scope.Scope, expr *parser.BinaryExpr, lut binopsLUT) types.Type {
	leftType := checkExpr(s, expr.Left)
	rightType := checkExpr(s, expr.Right)

	if leftType.IsError() || rightType.IsError() {
		return types.Error{}
	}

	if operLUT, ok := lut[expr.Oper]; ok {
		if leftLUT, ok := operLUT[leftType]; ok {
			if retType, ok := leftLUT[rightType]; ok {
				return retType
			}
		}

		msg := fmt.Sprintf("operator '%s' does not support %s and %s", expr.Oper, leftType, rightType)
		addTypeError(s, expr.Tok.Loc, msg)
		return types.Error{}
	}

	msg := fmt.Sprintf("unknown infix operator '%s'", expr.Oper)
	addTypeError(s, expr.Tok.Loc, msg)
	return types.Error{}
}

func checkListExpr(s scope.Scope, expr *parser.ListExpr) types.Type {
	var elemTypes []types.Type
	for _, elem := range expr.Elements {
		elemTypes = append(elemTypes, checkExpr(s, elem))
	}

	if len(elemTypes) == 0 {
		msg := "cannot determine type from empty list"
		addTypeError(s, expr.Start(), msg)
		return types.Error{}
	}

	for _, typ := range elemTypes {
		if typ.IsError() {
			return types.Error{}
		}
	}

	var listType types.Type
	for i, typ := range elemTypes {
		if listType == nil {
			listType = typ
			continue
		}

		if listType.Equals(typ) == false {
			msg := fmt.Sprintf("element type %s is not compatible with type %s", typ, listType)
			addTypeError(s, expr.Elements[i].Start(), msg)
			return types.Error{}
		}
	}

	return types.List{Child: listType}
}

func checkSubscriptExpr(s scope.Scope, expr *parser.SubscriptExpr, lut binopsLUT) types.Type {
	listType := checkExpr(s, expr.ListLike)
	indexType := checkExpr(s, expr.Index)

	if listType.IsError() || indexType.IsError() {
		return types.Error{}
	}

	if listType, ok := listType.(types.List); ok {
		return types.Optional{Child: listType.Child}
	}

	if subscriptLUT, ok := lut["["]; ok {
		if listLUT, ok := subscriptLUT[listType]; ok {
			if retType, ok := listLUT[indexType]; ok {
				return retType
			}
		}

		msg := fmt.Sprintf("subscript operator does not support %s[%s]", listType, indexType)
		addTypeError(s, expr.Index.Start(), msg)
		return types.Error{}
	}

	addTypeError(s, expr.Start(), "unknown infix operator '['")
	return types.Error{}
}

func checkSelfExpr(s scope.Scope, expr *parser.SelfExpr) types.Type {
	if s.HasSelfReference() == false {
		addTypeError(s, expr.Start(), "self references must be inside a function")
		return types.Error{}
	}

	return s.GetSelfReference()
}

func checkIdentExpr(s scope.Scope, expr *parser.IdentExpr) types.Type {
	if s.HasVariable(expr.Name) {
		return s.GetVariableType(expr.Name)
	}

	msg := fmt.Sprintf("variable '%s' was used before it was declared", expr.Name)
	addTypeError(s, expr.Start(), msg)
	return types.Error{}
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

// TypeCheckError combines a source code location with the resulting error message
type TypeCheckError struct {
	Loc     lexer.Loc
	Message string
}

func addTypeError(s scope.Scope, loc lexer.Loc, msg string) {
	err := TypeCheckError{loc, msg}
	s.NewError(err)
}

func (err TypeCheckError) Error() string {
	return fmt.Sprintf("%s %s", err.Loc, err.Message)
}
