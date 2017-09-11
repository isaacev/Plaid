package check

import (
	"fmt"
	"plaid/parser"
	"sort"
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
	global := makeScope(nil)
	checkProgram(global, prog)
	return global
}

// Scope tracks the symbol table and other data used during the check
type Scope struct {
	parent        *Scope
	pendingReturn Type
	variables     map[string]Type
	Errs          []error
	queue         []queuedFunctionBody
}

func (s *Scope) hasParent() bool {
	return (s.parent != nil)
}

func (s *Scope) registerVariable(name string, typ Type) {
	s.variables[name] = typ
}

func (s *Scope) hasVariable(name string) bool {
	_, exists := s.variables[name]
	return exists
}

func (s *Scope) getVariable(name string) Type {
	return s.variables[name]
}

func (s *Scope) addError(err error) {
	if s.hasParent() {
		s.parent.addError(err)
	} else {
		s.Errs = append(s.Errs, err)
	}
}

func (s *Scope) String() string {
	names := []string{}
	for name := range s.variables {
		names = append(names, name)
	}
	sort.Strings(names)

	out := "+----------+--------------+\n"
	out += "| Var      | Type         |\n"
	out += "| -------- | ------------ |\n"
	for _, name := range names {
		out += fmt.Sprintf("| %-8s | %-12s |\n", name, s.variables[name])
	}
	out += "+----------+--------------+\n"
	return out
}

type queuedFunctionBody struct {
	ret  Type
	expr parser.FunctionExpr
}

func (s *Scope) hasFunctionBodyQueue() bool {
	return len(s.queue) > 0
}

func (s *Scope) enqueueFunctionBody(ret Type, expr parser.FunctionExpr) {
	s.queue = append(s.queue, queuedFunctionBody{ret, expr})
}

func (s *Scope) dequeueFunctionBody() (Type, parser.FunctionExpr) {
	dequeued := s.queue[0]
	s.queue = s.queue[1:]
	return dequeued.ret, dequeued.expr
}

func makeScope(parent *Scope) *Scope {
	scope := &Scope{
		parent,
		nil,
		make(map[string]Type),
		[]error{},
		[]queuedFunctionBody{},
	}

	return scope
}

func makeScopePendingReturn(parent *Scope, ret Type) *Scope {
	scope := &Scope{
		parent,
		ret,
		make(map[string]Type),
		[]error{},
		[]queuedFunctionBody{},
	}

	return scope
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

	if scope.hasFunctionBodyQueue() {
		sig, expr := scope.dequeueFunctionBody()
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
	scope.enqueueFunctionBody(ret, expr)
	return sig
}

func checkFunctionBody(scope *Scope, ret Type, expr parser.FunctionExpr) {
	pushed := makeScopePendingReturn(scope, ret)
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
		scope.addError(fmt.Errorf("expected %d arguments, got %d", len(calleeFunc.params.children), len(argTypes)))
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
