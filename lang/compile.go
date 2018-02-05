package lang

import "fmt"

func Compile(mod Module) Bytecode {
	if mod.IsNative() == false {
		btc := compileModule(mod.(*ModuleVirtual))
		mod.(*ModuleVirtual).bytecode = &btc
		return btc
	}

	return Bytecode{}
}

func compileModule(mod *ModuleVirtual) (blob Bytecode) {
	for name, _ := range mod.scope.Local {
		blob.write(InstrReserve{name})
	}
	blob.append(compileRootStmts(mod, mod.structure.Stmts))
	blob.write(InstrHalt{})
	return blob
}

func compileRootStmts(mod *ModuleVirtual, stmts []Stmt) (blob Bytecode) {
	for _, stmt := range stmts {
		blob.append(compileRootStmt(mod, stmt))
	}
	return blob
}

func compileStmts(s *Scope, stmts []Stmt) (blob Bytecode) {
	for _, stmt := range stmts {
		blob.append(compileStmt(s, stmt))
	}
	return blob
}

func compileRootStmt(mod *ModuleVirtual, stmt Stmt) Bytecode {
	switch stmt := stmt.(type) {
	case *UseStmt:
		return compileUseStmt(mod, stmt)
	default:
		return compileStmt(mod.scope, stmt)
	}
}

func compileUseStmt(mod *ModuleVirtual, stmt *UseStmt) Bytecode {
	blob := compileStringExpr(mod.scope, stmt.Path)
	blob.write(InstrLoadMod{})
	return blob
}

func compileStmt(s *Scope, stmt Stmt) Bytecode {
	switch stmt := stmt.(type) {
	case *PubStmt:
		return compilePubStmt(s, stmt)
	case *IfStmt:
		return compileIfStmt(s, stmt)
	case *DeclarationStmt:
		return compileDeclarationStmt(s, stmt)
	case *ReturnStmt:
		return compileReturnStmt(s, stmt)
	case *ExprStmt:
		return compileExprStmt(s, stmt)
	default:
		panic(fmt.Sprintf("cannot compile %T", stmt))
	}
}

func compilePubStmt(s *Scope, stmt *PubStmt) Bytecode {
	// TODO: Unclear how to `pub` objects to other environments.
	return compileStmt(s, stmt.Stmt)
}

func compileIfStmt(s *Scope, stmt *IfStmt) Bytecode {
	blob := compileExpr(s, stmt.Cond)
	jump := blob.write(InstrNOP{}) // Pending jump to end of if-clause
	done := blob.append(compileStmts(s, stmt.Clause.Stmts))
	blob.overwrite(jump, InstrJumpFalse{done})
	return blob
}

func compileDeclarationStmt(s *Scope, stmt *DeclarationStmt) Bytecode {
	blob := compileExpr(s, stmt.Expr)
	name := stmt.Name.Name
	blob.write(InstrStore{name})
	return blob
}

func compileReturnStmt(s *Scope, stmt *ReturnStmt) (blob Bytecode) {
	if stmt.Expr == nil {
		blob.write(InstrPush{&ObjectNone{}})
	} else {
		blob.append(compileExpr(s, stmt.Expr))
	}
	blob.write(InstrReturn{})
	return blob
}

func compileExprStmt(s *Scope, stmt *ExprStmt) Bytecode {
	blob := compileExpr(s, stmt.Expr)
	blob.write(InstrPop{})
	return blob
}

func compileExpr(s *Scope, expr Expr) Bytecode {
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
	case *IdentExpr:
		return compileIdentExpr(s, expr)
	case *SelfExpr:
		return compileSelfExpr(s, expr)
	case *NumberExpr:
		return compileNumberExpr(s, expr)
	case *StringExpr:
		return compileStringExpr(s, expr)
	case *BooleanExpr:
		return compileBoolExpr(s, expr)
	default:
		panic(fmt.Sprintf("cannot compile %T", expr))
	}
}

func compileFunctionExpr(s *Scope, expr *FunctionExpr) (blob Bytecode) {
	local := s.Children[expr]
	var params []string
	for _, param := range expr.Params {
		name := param.Name.Name
		params = append(params, name)
	}

	blobBody := Bytecode{}
	for name := range local.Local {
		isParam := false
		for _, param := range expr.Params {
			if param.Name.Name == name {
				isParam = true
				break
			}
		}

		if isParam == false {
			blobBody.write(InstrReserve{name})
		}
	}

	blobBody.append(compileStmts(local, expr.Block.Stmts))
	blobBody.write(InstrPush{&ObjectNone{}})
	blobBody.write(InstrReturn{})

	function := &ObjectFunction{
		params:   params,
		bytecode: blobBody,
	}

	blob.write(InstrPush{function})
	blob.write(InstrCreateClosure{})
	return blob
}

func compileDispatchExpr(s *Scope, expr *DispatchExpr) (blob Bytecode) {
	for i := len(expr.Args) - 1; i >= 0; i-- {
		blob.append(compileExpr(s, expr.Args[i]))
	}
	blob.append(compileExpr(s, expr.Callee))
	blob.write(InstrDispatch{len(expr.Args)})
	return blob
}

func compileAssignExpr(s *Scope, expr *AssignExpr) Bytecode {
	blob := compileExpr(s, expr.Right)
	name := expr.Left.Name
	blob.write(InstrCopy{})
	blob.write(InstrStore{name})
	return blob
}

func compileBinaryExpr(s *Scope, expr *BinaryExpr) Bytecode {
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
		panic(fmt.Sprintf("cannot compile %T", expr))
	}
	return blob
}

func compileAccessExpr(s *Scope, expr *AccessExpr) Bytecode {
	blob := compileExpr(s, expr.Left)
	blob.write(InstrLoadAttr{expr.Right.(*IdentExpr).Name})
	return blob
}

func compileIdentExpr(s *Scope, expr *IdentExpr) (blob Bytecode) {
	blob.write(InstrLoad{expr.Name})
	return blob
}

func compileSelfExpr(s *Scope, expr *SelfExpr) (blob Bytecode) {
	blob.write(InstrLoadSelf{})
	return blob
}

func compileNumberExpr(s *Scope, expr *NumberExpr) (blob Bytecode) {
	blob.write(InstrPush{&ObjectInt{int64(expr.Val)}})
	return blob
}

func compileStringExpr(s *Scope, expr *StringExpr) (blob Bytecode) {
	blob.write(InstrPush{&ObjectStr{expr.Val}})
	return blob
}

func compileBoolExpr(s *Scope, expr *BooleanExpr) (blob Bytecode) {
	blob.write(InstrPush{&ObjectBool{expr.Val}})
	return blob
}
