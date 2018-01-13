package lang

import (
	"fmt"
)

type binopsLUT map[string]map[Type]map[Type]Type
type doubleLUT map[Type]map[Type]Type
type singleLUT map[Type]Type

var defaultBinopsLUT = binopsLUT{
	"+": doubleLUT{
		TypeNativeInt: singleLUT{TypeNativeInt: TypeNativeInt},
		TypeNativeStr: singleLUT{TypeNativeStr: TypeNativeStr},
	},
	"-": doubleLUT{
		TypeNativeInt: singleLUT{TypeNativeInt: TypeNativeInt},
	},
	"*": doubleLUT{
		TypeNativeInt: singleLUT{TypeNativeInt: TypeNativeInt},
	},
	"<": doubleLUT{
		TypeNativeInt: singleLUT{TypeNativeInt: TypeNativeBool},
	},
	"<=": doubleLUT{
		TypeNativeInt: singleLUT{TypeNativeInt: TypeNativeBool},
	},
	">": doubleLUT{
		TypeNativeInt: singleLUT{TypeNativeInt: TypeNativeBool},
	},
	">=": doubleLUT{
		TypeNativeInt: singleLUT{TypeNativeInt: TypeNativeBool},
	},
	"[": doubleLUT{
		TypeNativeStr: singleLUT{TypeNativeInt: TypeOptional{Child: TypeNativeStr}},
	},
}

// checkModule takes an existing abstract syntax tree and performs type checks and
// other correctness checks. It returns a list of any errors that were
// discovered inside the AST
func checkModule(root *VirtualModule, builtins ...Module) Module {
	root.scope = MakeGlobalScope()

	for _, mod := range builtins {
		root.scope.AddImport(mod.Scope())
	}

	for _, mod := range root.imports {
		root.scope.AddImport(mod.Scope())
	}

	checkProgram(root.scope, root.ast)
	return root
}

func checkProgram(s *GlobalScope, ast *RootNode) Scope {
	for _, stmt := range ast.Stmts {
		checkStmt(s, stmt)
	}

	return s
}

func checkStmt(s Scope, stmt Stmt) {
	switch stmt := stmt.(type) {
	case *PubStmt:
		checkPubStmt(s, stmt)
		break
	case *IfStmt:
		checkIfStmt(s, stmt)
		break
	case *DeclarationStmt:
		checkDeclarationStmt(s, stmt)
		break
	case *ReturnStmt:
		checkReturnStmt(s, stmt)
		break
	case *ExprStmt:
		checkExprAllowVoid(s, stmt.Expr)
		break
	}
}

func checkStmtBlock(s Scope, block *StmtBlock) {
	for _, stmt := range block.Stmts {
		checkStmt(s, stmt)
	}
}

func checkPubStmt(s Scope, stmt *PubStmt) {
	checkStmt(s, stmt.Stmt)

	var g *GlobalScope
	var ok bool
	if g, ok = s.(*GlobalScope); ok == false {
		addTypeError(s, stmt.Start(), "pub statement must be a top-level statement")
		return
	}

	name := stmt.Stmt.Name.Name
	typ := g.GetVariableType(name)
	g.newExport(name, typ)
}

func checkIfStmt(s Scope, stmt *IfStmt) {
	typ := checkExpr(s, stmt.Cond)
	if TypeNativeBool.Equals(typ) == false {
		addTypeError(s, stmt.Cond.Start(), "condition must resolve to a boolean")
	}

	checkStmtBlock(s, stmt.Clause)
}

func checkDeclarationStmt(s Scope, stmt *DeclarationStmt) {
	name := stmt.Name.Name
	typ := checkExpr(s, stmt.Expr)
	s.newVariable(name, typ)
}

func checkReturnStmt(s Scope, stmt *ReturnStmt) {
	var ret Type = TypeVoid{}
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

	if s.GetSelfReference().Ret.Equals(TypeVoid{}) {
		msg := fmt.Sprintf("expected to return nothing, got '%s'", ret)
		addTypeError(s, stmt.Expr.Start(), msg)
		return
	}

	if (TypeVoid{}).Equals(ret) {
		msg := fmt.Sprintf("expected a return type of '%s', got nothing", s.GetSelfReference().Ret)
		addTypeError(s, stmt.Start(), msg)
		return
	}

	msg := fmt.Sprintf("expected to return '%s', got '%s'", s.GetSelfReference().Ret, ret)
	addTypeError(s, stmt.Expr.Start(), msg)
}

func checkExprAllowVoid(s Scope, expr Expr) Type {
	var typ Type = TypeError{}
	switch expr := expr.(type) {
	case *FunctionExpr:
		typ = checkFunctionExpr(s, expr)
	case *DispatchExpr:
		typ = checkDispatchExpr(s, expr)
	case *AssignExpr:
		typ = checkAssignExpr(s, expr)
	case *BinaryExpr:
		typ = checkBinaryExpr(s, expr, defaultBinopsLUT)
	case *ListExpr:
		typ = checkListExpr(s, expr)
	case *SubscriptExpr:
		typ = checkSubscriptExpr(s, expr, defaultBinopsLUT)
	case *SelfExpr:
		typ = checkSelfExpr(s, expr)
	case *IdentExpr:
		typ = checkIdentExpr(s, expr)
	case *NumberExpr:
		typ = checkNumberExpr(s, expr)
	case *StringExpr:
		typ = checkStringExpr(s, expr)
	case *BooleanExpr:
		typ = checkBooleanExpr(s, expr)
	default:
		addTypeError(s, expr.Start(), "unknown expression type")
	}

	return typ
}

func checkExpr(s Scope, expr Expr) Type {
	typ := checkExprAllowVoid(s, expr)

	if (TypeVoid{}).Equals(typ) {
		addTypeError(s, expr.Start(), "cannot use void types in an expression")
		return TypeError{}
	}

	return typ
}

func checkFunctionExpr(s Scope, expr *FunctionExpr) Type {
	ret := convertTypeNote(expr.Ret)
	params := []Type{}
	for _, param := range expr.Params {
		params = append(params, convertTypeNote(param.Note))
	}
	tuple := TypeTuple{Children: params}
	self := TypeFunction{Params: tuple, Ret: ret}

	childScope := makeLocalScope(s, self)

	for _, param := range expr.Params {
		paramName := param.Name.Name
		paramType := convertTypeNote(param.Note)
		childScope.newVariable(paramName, paramType)
	}

	checkStmtBlock(childScope, expr.Block)
	return self
}

func checkDispatchExpr(s Scope, expr *DispatchExpr) Type {
	// Resolve arguments to types
	argTypes := []Type{}
	for _, argExpr := range expr.Args {
		argTypes = append(argTypes, checkExpr(s, argExpr))
	}

	// Resolve callee to type
	calleeType := checkExpr(s, expr.Callee)
	calleeFunc, ok := calleeType.(TypeFunction)
	if ok == false {
		if calleeType.IsError() == false {
			msg := fmt.Sprintf("cannot call function on type '%s'", calleeType)
			addTypeError(s, expr.Start(), msg)
		}

		return TypeError{}
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
				retType = TypeError{}
			} else if paramType.Equals(argType) == false {
				msg := fmt.Sprintf("expected '%s', got '%s'", paramType, argType)
				addTypeError(s, expr.Args[i].Start(), msg)
				retType = TypeError{}
			}
		}
	} else {
		msg := fmt.Sprintf("expected %d arguments, got %d", totalParams, totalArgs)
		addTypeError(s, expr.Start(), msg)
		retType = TypeError{}
	}

	return retType
}

func checkAssignExpr(s Scope, expr *AssignExpr) Type {
	name := expr.Left.Name
	leftType := s.GetVariableType(name)
	rightType := checkExpr(s, expr.Right)

	if leftType == nil {
		msg := fmt.Sprintf("'%s' cannot be assigned before it is declared", name)
		addTypeError(s, expr.Start(), msg)
		return TypeError{}
	}

	if leftType.IsError() || rightType.IsError() {
		return TypeError{}
	}

	if leftType.Equals(rightType) == false {
		msg := fmt.Sprintf("'%s' cannot be assigned type '%s'", leftType, rightType)
		addTypeError(s, expr.Right.Start(), msg)
		return TypeError{}
	}

	return leftType
}

func checkBinaryExpr(s Scope, expr *BinaryExpr, lut binopsLUT) Type {
	leftType := checkExpr(s, expr.Left)
	rightType := checkExpr(s, expr.Right)

	if leftType.IsError() || rightType.IsError() {
		return TypeError{}
	}

	if operLUT, ok := lut[expr.Oper]; ok {
		if leftLUT, ok := operLUT[leftType]; ok {
			if retType, ok := leftLUT[rightType]; ok {
				return retType
			}
		}

		msg := fmt.Sprintf("operator '%s' does not support %s and %s", expr.Oper, leftType, rightType)
		addTypeError(s, expr.Tok.Loc, msg)
		return TypeError{}
	}

	msg := fmt.Sprintf("unknown infix operator '%s'", expr.Oper)
	addTypeError(s, expr.Tok.Loc, msg)
	return TypeError{}
}

func checkListExpr(s Scope, expr *ListExpr) Type {
	var elemTypes []Type
	for _, elem := range expr.Elements {
		elemTypes = append(elemTypes, checkExpr(s, elem))
	}

	if len(elemTypes) == 0 {
		msg := "cannot determine type from empty list"
		addTypeError(s, expr.Start(), msg)
		return TypeError{}
	}

	for _, typ := range elemTypes {
		if typ.IsError() {
			return TypeError{}
		}
	}

	var listType Type
	for i, typ := range elemTypes {
		if listType == nil {
			listType = typ
			continue
		}

		if listType.Equals(typ) == false {
			msg := fmt.Sprintf("element type %s is not compatible with type %s", typ, listType)
			addTypeError(s, expr.Elements[i].Start(), msg)
			return TypeError{}
		}
	}

	return TypeList{Child: listType}
}

func checkSubscriptExpr(s Scope, expr *SubscriptExpr, lut binopsLUT) Type {
	listType := checkExpr(s, expr.ListLike)
	indexType := checkExpr(s, expr.Index)

	if listType.IsError() || indexType.IsError() {
		return TypeError{}
	}

	if listType, ok := listType.(TypeList); ok {
		return TypeOptional{Child: listType.Child}
	}

	if subscriptLUT, ok := lut["["]; ok {
		if listLUT, ok := subscriptLUT[listType]; ok {
			if retType, ok := listLUT[indexType]; ok {
				return retType
			}
		}

		msg := fmt.Sprintf("subscript operator does not support %s[%s]", listType, indexType)
		addTypeError(s, expr.Index.Start(), msg)
		return TypeError{}
	}

	addTypeError(s, expr.Start(), "unknown infix operator '['")
	return TypeError{}
}

func checkSelfExpr(s Scope, expr *SelfExpr) Type {
	if s.HasSelfReference() == false {
		addTypeError(s, expr.Start(), "self references must be inside a function")
		return TypeError{}
	}

	return s.GetSelfReference()
}

func checkIdentExpr(s Scope, expr *IdentExpr) Type {
	if s.HasVariable(expr.Name) {
		return s.GetVariableType(expr.Name)
	}

	msg := fmt.Sprintf("variable '%s' was used before it was declared", expr.Name)
	addTypeError(s, expr.Start(), msg)
	return TypeError{}
}

func checkNumberExpr(s Scope, expr *NumberExpr) Type {
	return TypeNativeInt
}

func checkStringExpr(s Scope, expr *StringExpr) Type {
	return TypeNativeStr
}

func checkBooleanExpr(s Scope, expr *BooleanExpr) Type {
	return TypeNativeBool
}

// TypeCheckError combines a source code location with the resulting error message
type TypeCheckError struct {
	Loc     Loc
	Message string
}

func addTypeError(s Scope, loc Loc, msg string) {
	err := TypeCheckError{loc, msg}
	s.newError(err)
}

func (err TypeCheckError) Error() string {
	return fmt.Sprintf("%s %s", err.Loc, err.Message)
}
