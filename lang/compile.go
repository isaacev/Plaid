package lang

/*
import "fmt"

func Compile(mod Module) Bytecode {
	if virt, ok := mod.(*VirtualModule); ok {
		main := compileProgram(mod, virt.Scope(), virt.ast)
		return main
	}

	return Bytecode{}
}

func compileProgram(mod Module, s *GlobalScope, prog *RootNode) Bytecode {
	blob := Bytecode{}
	for _, name := range s.GetLocalVariableNames() {
		symbol := s.GetLocalVariableReference(name)
		blob.write(InstrReserve{name, symbol})
	}

	blob.append(compileTopLevelStmts(mod, s, prog.Stmts))
	blob.write(InstrHalt{})
	return blob
}

func compileTopLevelStmts(mod Module, s *GlobalScope, stmts []Stmt) (blob Bytecode) {
	for _, stmt := range stmts {
		blob.append(compileTopLevelStmt(mod, s, stmt))
	}
	return blob
}

func compileTopLevelStmt(mod Module, s *GlobalScope, stmt Stmt) Bytecode {
	switch stmt := stmt.(type) {
	case *UseStmt:
		return compileUseStmt(mod, s, stmt)
	default:
		return compileStmt(s, stmt)
	}
}

func compileStmts(s Scope, stmts []Stmt) (blob Bytecode) {
	for _, stmt := range stmts {
		blob.append(compileStmt(s, stmt))
	}
	return blob
}

func compileStmt(s Scope, stmt Stmt) Bytecode {
	switch stmt := stmt.(type) {
	case *PubStmt:
		return compilePubStmt(s, stmt)
	case *IfStmt:
		return compileIfStmt(s, stmt)
	case *ReturnStmt:
		return compileReturnStmt(s, stmt)
	case *DeclarationStmt:
		return compileDeclarationStmt(s, stmt)
	case *ExprStmt:
		return compileExprStmt(s, stmt)
	default:
		panic(fmt.Sprintf("cannot compile %T", stmt))
	}
}

func compileUseStmt(mod Module, s *GlobalScope, stmt *UseStmt) Bytecode {
	name := stmt.Path.Val
	for _, imp := range mod.Imports() {
		if imp.Path() == name {
			if native, ok := imp.(*NativeModule); ok {
				lib := native.library
				blob := Bytecode{}
				blob.write(InstrPush{lib.toObject()})
				blob.write(InstrStore{name, s.GetVariableReference(name)})
				return blob
			}
		}
	}
	panic(fmt.Sprintf("cannot find library named '%s'", stmt.Path.Val))
}

func compilePubStmt(s Scope, stmt *PubStmt) Bytecode {
	return compileStmt(s, stmt.Stmt)
}

func compileIfStmt(s Scope, stmt *IfStmt) Bytecode {
	blob := compileExpr(s, stmt.Cond)
	jump := blob.write(InstrNOP{}) // Pending jump to end of clause
	done := blob.append(compileStmts(s, stmt.Clause.Stmts))
	blob.overwrite(jump, InstrJumpFalse{done})
	return blob
}

func compileReturnStmt(s Scope, stmt *ReturnStmt) (blob Bytecode) {
	if stmt.Expr == nil {
		blob.write(InstrPush{ObjectNone{}})
	} else {
		blob.append(compileExpr(s, stmt.Expr))
	}
	blob.write(InstrReturn{})
	return blob
}

func compileDeclarationStmt(s Scope, stmt *DeclarationStmt) Bytecode {
	blob := compileExpr(s, stmt.Expr)
	symbol := s.GetVariableReference(stmt.Name.Name)
	blob.write(InstrStore{stmt.Name.Name, symbol})
	return blob
}

func compileExprStmt(s Scope, stmt *ExprStmt) Bytecode {
	blob := compileExpr(s, stmt.Expr)
	blob.write(InstrPop{})
	return blob
}

func compileExpr(s Scope, expr Expr) Bytecode {
	switch expr := expr.(type) {
	case *FunctionExpr:
		return compileFunctionExpr(s, expr)
	case *DispatchExpr:
		return compileDispatchExpr(s, expr)
	case *AssignExpr:
		return compileAssignExpr(s, expr)
	case *BinaryExpr:
		return compileBinaryExpr(s, expr)
	case *AccessExpr:
		return compileAccessExpr(s, expr)
	case *SelfExpr:
		return compileSelfExpr(s, expr)
	case *IdentExpr:
		return compileIdentExpr(s, expr)
	case *NumberExpr:
		return compileNumberExpr(s, expr)
	case *StringExpr:
		return compileStringExpr(s, expr)
	case *BooleanExpr:
		return compileBoolExpr(s, expr)
	default:
		panic(fmt.Sprintf("cannot transform expression %T", expr))
	}
}

func compileFunctionExpr(s Scope, expr *FunctionExpr) (blob Bytecode) {
	local := s.GetChild(expr)
	var params []*UniqueSymbol
	for _, param := range expr.Params {
		name := param.Name.Name
		symbol := local.GetLocalVariableReference(name)
		params = append(params, symbol)
	}

	bodyBlob := Bytecode{}
	for _, name := range local.GetLocalVariableNames() {
		isParam := false
		for _, param := range expr.Params {
			if param.Name.Name == name {
				isParam = true
				break
			}
		}

		if isParam == false {
			symbol := local.GetLocalVariableReference(name)
			bodyBlob.write(InstrReserve{name, symbol})
		}
	}

	bodyBlob.append(compileStmts(local, expr.Block.Stmts))
	bodyBlob.write(InstrPush{ObjectNone{}})
	bodyBlob.write(InstrReturn{})

	function := ObjectFunction{
		params:   params,
		bytecode: bodyBlob,
	}

	blob.write(InstrPush{function})
	return blob
}

func compileDispatchExpr(s Scope, expr *DispatchExpr) (blob Bytecode) {
	for i := len(expr.Args) - 1; i >= 0; i-- {
		blob.append(compileExpr(s, expr.Args[i]))
	}
	blob.append(compileExpr(s, expr.Callee))
	blob.write(InstrDispatch{args: len(expr.Args)})
	return blob
}

func compileAccessExpr(s Scope, expr *AccessExpr) Bytecode {
	blob := compileExpr(s, expr.Left)
	blob.write(InstrLoadAttr{expr.Right.(*IdentExpr).Name})
	return blob
}

func compileAssignExpr(s Scope, expr *AssignExpr) Bytecode {
	blob := compileExpr(s, expr.Right)
	symbol := s.GetVariableReference(expr.Left.Name)
	blob.write(InstrCopy{})
	blob.write(InstrStore{expr.Left.Name, symbol})
	return blob
}

func compileBinaryExpr(s Scope, expr *BinaryExpr) Bytecode {
	blob := compileExpr(s, expr.Left)
	blob.append(compileExpr(s, expr.Right))

	switch expr.Oper {
	case "+":
		blob.write(InstrAdd{})
	case "-":
		blob.write(InstrSub{})
	case "*":
		blob.write(InstrMul{})
	case "<":
		blob.write(InstrLT{})
	case "<=":
		blob.write(InstrLTEquals{})
	case ">":
		blob.write(InstrGT{})
	case ">=":
		blob.write(InstrGTEquals{})
	default:
		panic(fmt.Sprintf("cannot compile %T with '%s'", expr, expr.Oper))
	}

	return blob
}

func compileSelfExpr(s Scope, expr *SelfExpr) Bytecode {
	blob := Bytecode{}
	blob.write(InstrLoadSelf{})
	return blob
}

func compileIdentExpr(s Scope, expr *IdentExpr) Bytecode {
	blob := Bytecode{}
	symbol := s.GetVariableReference(expr.Name)

	if symbol == nil {
		panic("nil lookup")
	}

	blob.write(InstrLoad{expr.Name, symbol})
	return blob
}

func compileNumberExpr(s Scope, expr *NumberExpr) (blob Bytecode) {
	blob.write(InstrPush{ObjectInt{int64(expr.Val)}})
	return blob
}

func compileStringExpr(s Scope, expr *StringExpr) (blob Bytecode) {
	blob.write(InstrPush{ObjectStr{expr.Val}})
	return blob
}

func compileBoolExpr(s Scope, expr *BooleanExpr) (blob Bytecode) {
	blob.write(InstrPush{ObjectBool{expr.Val}})
	return blob
}
*/
