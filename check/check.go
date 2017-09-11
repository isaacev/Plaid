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

// Scope tracks the symbol table and other data used during the check
type Scope struct {
	parent        *Scope
	errs          []error
	variables     []string
	values        map[string]Type
	pendingReturn Type
	queue         []struct {
		ret  Type
		expr parser.FunctionExpr
	}
}

func (s *Scope) hasParent() bool {
	return (s.parent != nil)
}

func (s *Scope) addError(err error) {
	if s.hasParent() {
		s.parent.addError(err)
	} else {
		s.errs = append(s.errs, err)
	}
}

// Errors returns a list of errors detected during the check
func (s *Scope) Errors() []error {
	return s.errs
}

func (s *Scope) hasVariable(name string) bool {
	_, exists := s.values[name]
	return exists
}

func (s *Scope) registerVariable(name string, typ Type) {
	s.variables = append(s.variables, name)
	s.values[name] = typ
}

func (s *Scope) getVariable(name string) Type {
	return s.values[name]
}

func (s *Scope) hasPendingReturnType() bool {
	return (s.pendingReturn != nil)
}

func (s *Scope) getPendingReturnType() Type {
	return s.pendingReturn
}

func (s *Scope) setPendingReturnType(typ Type) {
	s.pendingReturn = typ
}

func (s *Scope) hasBodyQueue() bool {
	return len(s.queue) > 0
}

func (s *Scope) enqueueBody(ret Type, expr parser.FunctionExpr) {
	body := struct {
		ret  Type
		expr parser.FunctionExpr
	}{ret, expr}
	s.queue = append(s.queue, body)
}

func (s *Scope) dequeueBody() (Type, parser.FunctionExpr) {
	body := s.queue[0]
	s.queue = s.queue[1:]
	return body.ret, body.expr
}

func (s *Scope) String() string {
	var out string
	for _, name := range s.variables {
		out += fmt.Sprintf("%s : %s\n", name, s.values[name])
	}
	return out
}

func makeScope(parent *Scope, ret Type) *Scope {
	return &Scope{
		parent,
		[]error{},
		[]string{},
		make(map[string]Type),
		ret,
		[]struct {
			ret  Type
			expr parser.FunctionExpr
		}{},
	}
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
	scope.registerVariable(name, typ)
}

func checkReturnStmt(scope *Scope, stmt parser.ReturnStmt) {
	var ret Type
	if stmt.Expr != nil {
		ret = checkExpr(scope, stmt.Expr)
	}

	expectedReturnValue := scope.pendingReturn != nil
	gotReturnValue := ret != nil

	if scope.hasParent() {
		if expectedReturnValue {
			if gotReturnValue == false {
				scope.addError(fmt.Errorf("expected a return type of '%s', got nothing", scope.pendingReturn))
			} else if scope.pendingReturn.Equals(ret) == false {
				scope.addError(fmt.Errorf("expected to return '%s', got '%s'", scope.pendingReturn, ret))
			}
		} else {
			if gotReturnValue {
				scope.addError(fmt.Errorf("expected to return nothing, got '%s'", ret))
			}
		}
	} else {
		scope.addError(fmt.Errorf("return statements must be inside a function"))
	}
}

func checkExpr(scope *Scope, expr parser.Expr) Type {
	switch expr := expr.(type) {
	case parser.FunctionExpr:
		return checkFunctionExpr(scope, expr)
	case parser.DispatchExpr:
		return checkDispatchExpr(scope, expr)
	case parser.BinaryExpr:
		return checkBinaryExpr(scope, expr)
	case parser.IdentExpr:
		return checkIdentExpr(scope, expr)
	case parser.NumberExpr:
		return checkNumberExpr(scope, expr)
	case parser.StringExpr:
		return checkStringExpr(scope, expr)
	default:
		scope.addError(fmt.Errorf("unknown expression type"))
		return TypeError{}
	}
}

func checkFunctionExpr(scope *Scope, expr parser.FunctionExpr) Type {
	ret := convertTypeSig(expr.Ret)
	params := []Type{}
	for _, param := range expr.Params {
		params = append(params, convertTypeSig(param.Sig))
	}
	tuple := TypeTuple{params}

	sig := TypeFunction{tuple, ret}
	scope.enqueueBody(ret, expr)
	return sig
}

func checkFunctionBody(scope *Scope, ret Type, expr parser.FunctionExpr) {
	pushed := makeScope(scope, ret)
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

func checkBinaryExpr(scope *Scope, expr parser.BinaryExpr) Type {
	switch expr.Oper {
	case "+":
		return checkAddition(scope, expr.Left, expr.Right)
	default:
		scope.addError(fmt.Errorf("unknown infix operator '%s'", expr.Oper))
		return TypeError{}
	}
}

func checkAddition(scope *Scope, left parser.Expr, right parser.Expr) Type {
	leftType := checkExpr(scope, left)
	rightType := checkExpr(scope, right)

	if leftType.IsError() || rightType.IsError() {
		return TypeError{}
	}

	typ := BuiltinInt

	if leftType.Equals(BuiltinInt) == false {
		scope.addError(fmt.Errorf("left side must have type %s, got %s", BuiltinInt, leftType))
		typ = TypeError{}
	}

	if rightType.Equals(BuiltinInt) == false {
		scope.addError(fmt.Errorf("right side must have type %s, got %s", BuiltinInt, rightType))
		typ = TypeError{}
	}

	return typ
}

func checkIdentExpr(scope *Scope, expr parser.IdentExpr) Type {
	if scope.hasVariable(expr.Name) {
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

func convertTypeSig(sig parser.TypeSig) Type {
	switch sig := sig.(type) {
	case parser.TypeFunction:
		return TypeFunction{convertTypeSig(sig.Params).(TypeTuple), convertTypeSig(sig.Ret)}
	case parser.TypeTuple:
		elems := []Type{}
		for _, elem := range sig.Elems {
			elems = append(elems, convertTypeSig(elem))
		}
		return TypeTuple{elems}
	case parser.TypeList:
		return TypeList{convertTypeSig(sig.Child)}
	case parser.TypeOptional:
		return TypeOptional{convertTypeSig(sig.Child)}
	case parser.TypeIdent:
		return TypeIdent{sig.Name}
	default:
		return nil
	}
}
