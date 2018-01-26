package lang

import (
	"fmt"
	"plaid/lang/types"
)

func Check(mod Module) (Module, []error) {
	switch cast := mod.(type) {
	case *NativeModule:
		return cast, nil
	default:
		panic("unknown module type")
	}
}

type binopsLUT map[string]map[types.Type]map[types.Type]types.Type
type doubleLUT map[types.Type]map[types.Type]types.Type
type singleLUT map[types.Type]types.Type

var defaultBinopsLUT = binopsLUT{
	"+": doubleLUT{
		types.BuiltinInt: singleLUT{types.BuiltinInt: types.BuiltinInt},
		types.BuiltinStr: singleLUT{types.BuiltinStr: types.BuiltinStr},
	},
	"-": doubleLUT{
		types.BuiltinInt: singleLUT{types.BuiltinInt: types.BuiltinInt},
	},
	"*": doubleLUT{
		types.BuiltinInt: singleLUT{types.BuiltinInt: types.BuiltinInt},
	},
	"<": doubleLUT{
		types.BuiltinInt: singleLUT{types.BuiltinInt: types.BuiltinBool},
	},
	"<=": doubleLUT{
		types.BuiltinInt: singleLUT{types.BuiltinInt: types.BuiltinBool},
	},
	">": doubleLUT{
		types.BuiltinInt: singleLUT{types.BuiltinInt: types.BuiltinBool},
	},
	">=": doubleLUT{
		types.BuiltinInt: singleLUT{types.BuiltinInt: types.BuiltinBool},
	},
	"[": doubleLUT{
		types.BuiltinStr: singleLUT{types.BuiltinInt: types.Optional{Child: types.BuiltinStr}},
	},
}

// checkModule takes as input a *VirtualModule to check for semantic errors. It
// returns the root of a scope tree that describes the lifecycles of all symbols
// within the program. The semantic analysis tries to find as many semantic
// errors as possible in a single pass. Any errors that are detected are
// available by calling `GetErrors()` on the returned scope object.
func checkModule(root *VirtualModule) *GlobalScope {
	global := makeGlobalScope()

	for _, mod := range root.imports {
		if mod.Scope() == nil {
			global.addImport(checkModule(mod.(*VirtualModule)))
		} else {
			global.addImport(mod.Scope())
		}
	}

	checkProgram(global, root.ast)
	root.scope = global
	return global
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
	if types.BuiltinBool.Equals(typ) == false {
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

func checkExprAllowVoid(s Scope, expr Expr) types.Type {
	var typ types.Type = types.Error{}
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

func checkExpr(s Scope, expr Expr) types.Type {
	typ := checkExprAllowVoid(s, expr)

	if (types.Void{}).Equals(typ) {
		addTypeError(s, expr.Start(), "cannot use void types in an expression")
		return types.Error{}
	}

	return typ
}

func checkFunctionExpr(s Scope, expr *FunctionExpr) types.Type {
	ret := convertTypeNote(expr.Ret)
	params := []types.Type{}
	for _, param := range expr.Params {
		params = append(params, convertTypeNote(param.Note))
	}
	tuple := types.Tuple{Children: params}
	self := types.Function{Params: tuple, Ret: ret}

	childScope := makeLocalScope(s, expr, self)

	for _, param := range expr.Params {
		paramName := param.Name.Name
		paramType := convertTypeNote(param.Note)
		childScope.newVariable(paramName, paramType)
	}

	checkStmtBlock(childScope, expr.Block)
	return self
}

func checkDispatchExpr(s Scope, expr *DispatchExpr) types.Type {
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

func checkAssignExpr(s Scope, expr *AssignExpr) types.Type {
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

func checkBinaryExpr(s Scope, expr *BinaryExpr, lut binopsLUT) types.Type {
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

func checkListExpr(s Scope, expr *ListExpr) types.Type {
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

func checkSubscriptExpr(s Scope, expr *SubscriptExpr, lut binopsLUT) types.Type {
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

func checkSelfExpr(s Scope, expr *SelfExpr) types.Type {
	if s.HasSelfReference() == false {
		addTypeError(s, expr.Start(), "self references must be inside a function")
		return types.Error{}
	}

	return s.GetSelfReference()
}

func checkIdentExpr(s Scope, expr *IdentExpr) types.Type {
	if s.HasVariable(expr.Name) {
		return s.GetVariableType(expr.Name)
	}

	msg := fmt.Sprintf("variable '%s' was used before it was declared", expr.Name)
	addTypeError(s, expr.Start(), msg)
	return types.Error{}
}

func checkNumberExpr(s Scope, expr *NumberExpr) types.Type {
	return types.BuiltinInt
}

func checkStringExpr(s Scope, expr *StringExpr) types.Type {
	return types.BuiltinStr
}

func checkBooleanExpr(s Scope, expr *BooleanExpr) types.Type {
	return types.BuiltinBool
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

// convertTypeNote transforms a TypeNote struct (used to represent a syntax
// type notation) into a Type struct (used internally to represent a type)
func convertTypeNote(note TypeNote) types.Type {
	switch note := note.(type) {
	case TypeNoteAny:
		return types.Any{}
	case TypeNoteVoid:
		return types.Void{}
	case TypeNoteFunction:
		return types.Function{
			Params: convertTypeNote(note.Params).(types.Tuple),
			Ret:    convertTypeNote(note.Ret),
		}
	case TypeNoteTuple:
		elems := []types.Type{}
		for _, elem := range note.Elems {
			elems = append(elems, convertTypeNote(elem))
		}
		return types.Tuple{Children: elems}
	case TypeNoteList:
		return types.List{Child: convertTypeNote(note.Child)}
	case TypeNoteOptional:
		return types.Optional{Child: convertTypeNote(note.Child)}
	case TypeNoteIdent:
		return types.Ident{Name: note.Name}
	default:
		return nil
	}
}
