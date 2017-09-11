package check

import (
	"fmt"
	"plaid/parser"
)

// A collection of types native to the execution environment
var (
	BuiltinInt Type = TypeIdent{"Int"}
	BuiltinStr      = TypeIdent{"Str"}
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
	case parser.ReturnStmt:
		checkReturnStmt(scope, stmt)
		break
	case parser.ExprStmt:
		checkExpr(scope, stmt.Expr)
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

func checkReturnStmt(scope *Scope, stmt parser.ReturnStmt) {
	var ret Type = TypeVoid{}
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

	if scope.pendingReturn.Equals(TypeVoid{}) {
		scope.addError(fmt.Errorf("expected to return nothing, got '%s'", ret))
		return
	}

	if ret.Equals(TypeVoid{}) {
		scope.addError(fmt.Errorf("expected a return type of '%s', got nothing", scope.pendingReturn))
	}

	scope.addError(fmt.Errorf("expected to return '%s', got '%s'", scope.pendingReturn, ret))
}

func checkExpr(scope *Scope, expr parser.Expr) Type {
	var typ Type = TypeError{}
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

func checkFunctionExpr(scope *Scope, expr parser.FunctionExpr) Type {
	ret := convertTypeNote(expr.Ret)
	params := []Type{}
	for _, param := range expr.Params {
		params = append(params, convertTypeNote(param.Note))
	}
	tuple := TypeTuple{params}

	sig := TypeFunction{tuple, ret}
	scope.enqueueBody(ret, expr)
	return sig
}

func checkFunctionBody(scope *Scope, ret Type, expr parser.FunctionExpr) {
	pushed := makeScope(scope, ret)

	for _, param := range expr.Params {
		paramName := param.Name.Name
		paramType := convertTypeNote(param.Note)
		pushed.registerLocalVariable(paramName, paramType)
	}

	checkStmtBlock(pushed, expr.Block)
}

func checkDispatchExpr(scope *Scope, expr parser.DispatchExpr) Type {
	// Resolve arguments to types
	argTypes := []Type{}
	for _, argExpr := range expr.Args {
		argTypes = append(argTypes, checkExpr(scope, argExpr))
	}

	// Resolve callee to type
	calleeType := checkExpr(scope, expr.Callee)
	calleeFunc, ok := calleeType.(TypeFunction)
	if ok == false {
		if calleeType.IsError() == false {
			scope.addError(fmt.Errorf("cannot call function on type '%s'", calleeType))
		}

		return TypeError{}
	}

	// Resolve return type
	retType := calleeFunc.ret

	// Check that the given argument types match the expected parameter types
	totalArgs := len(argTypes)
	totalParams := len(calleeFunc.params.children)
	if totalArgs == totalParams {
		for i := 0; i < totalArgs; i++ {
			argType := argTypes[i]
			paramType := calleeFunc.params.children[i]

			if argType.Equals(paramType) == false {
				scope.addError(fmt.Errorf("expected '%s', got '%s'", paramType, argType))
				retType = TypeError{}
			}
		}
	} else {
		scope.addError(fmt.Errorf("expected %d arguments, got %d", totalParams, totalArgs))
		retType = TypeError{}
	}

	return retType
}

func checkAssignExpr(scope *Scope, expr parser.AssignExpr) Type {
	name := expr.Left.Name
	leftType := scope.getVariable(name)
	rightType := checkExpr(scope, expr.Right)

	if leftType.IsError() || rightType.IsError() {
		return TypeError{}
	}

	if leftType.Equals(rightType) == false {
		scope.addError(fmt.Errorf("'%s' cannot be assigned type '%s'", leftType, rightType))
		return TypeError{}
	}

	return leftType
}

func checkBinaryExpr(scope *Scope, expr parser.BinaryExpr) Type {
	switch expr.Oper {
	case "+":
		return expectBinaryTypes(scope, expr.Left, BuiltinInt, expr.Right, BuiltinInt, BuiltinInt)
	case "-":
		return expectBinaryTypes(scope, expr.Left, BuiltinInt, expr.Right, BuiltinInt, BuiltinInt)
	default:
		scope.addError(fmt.Errorf("unknown infix operator '%s'", expr.Oper))
		return TypeError{}
	}
}

func expectBinaryTypes(scope *Scope, left parser.Expr, expLeftType Type, right parser.Expr, expRightType Type, retType Type) Type {
	leftType := checkExpr(scope, left)
	rightType := checkExpr(scope, right)

	if leftType.IsError() || rightType.IsError() {
		return TypeError{}
	}

	typ := retType

	if leftType.Equals(expLeftType) == false {
		scope.addError(fmt.Errorf("left side must have type %s, got %s", expLeftType, leftType))
		typ = TypeError{}
	}

	if rightType.Equals(expRightType) == false {
		scope.addError(fmt.Errorf("right side must have type %s, got %s", expRightType, rightType))
		typ = TypeError{}
	}

	return typ
}

func checkIdentExpr(scope *Scope, expr parser.IdentExpr) Type {
	if scope.existingVariable(expr.Name) {
		return scope.getVariable(expr.Name)
	}

	scope.addError(fmt.Errorf("variable '%s' was used before it was declared", expr.Name))
	return TypeError{}
}

func checkNumberExpr(scope *Scope, expr parser.NumberExpr) Type {
	return BuiltinInt
}

func checkStringExpr(scope *Scope, expr parser.StringExpr) Type {
	return BuiltinStr
}

func convertTypeNote(note parser.TypeNote) Type {
	switch note := note.(type) {
	case parser.TypeNoteVoid:
		return TypeVoid{}
	case parser.TypeNoteFunction:
		return TypeFunction{convertTypeNote(note.Params).(TypeTuple), convertTypeNote(note.Ret)}
	case parser.TypeNoteTuple:
		elems := []Type{}
		for _, elem := range note.Elems {
			elems = append(elems, convertTypeNote(elem))
		}
		return TypeTuple{elems}
	case parser.TypeNoteList:
		return TypeList{convertTypeNote(note.Child)}
	case parser.TypeNoteOptional:
		return TypeOptional{convertTypeNote(note.Child)}
	case parser.TypeNoteIdent:
		return TypeIdent{note.Name}
	default:
		return nil
	}
}
