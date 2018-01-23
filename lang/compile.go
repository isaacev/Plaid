package lang

import "fmt"

func Compile(mod Module) bytecode {
	if virt, ok := mod.(*VirtualModule); ok {
		main := compileProgram(virt.Scope(), virt.ast)
		fmt.Println(main.String())
		return main
	}

	return bytecode{}
}

func compileProgram(s Scope, prog *RootNode) bytecode {
	blob := bytecode{}
	for _, name := range s.GetLocalVariableNames() {
		symbol := s.GetLocalVariableReference(name)
		blob.write(InstrReserve{name, symbol})
	}

	blob.append(compileStmts(s, prog.Stmts))
	blob.write(InstrHalt{})
	return blob
}

func compileStmts(s Scope, stmts []Stmt) (blob bytecode) {
	for _, stmt := range stmts {
		blob.append(compileStmt(s, stmt))
	}
	return blob
}

func compileStmt(s Scope, stmt Stmt) bytecode {
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

func compilePubStmt(s Scope, stmt *PubStmt) bytecode {
	return compileStmt(s, stmt.Stmt)
}

func compileIfStmt(s Scope, stmt *IfStmt) bytecode {
	blob := compileExpr(s, stmt.Cond)
	jump := blob.write(InstrNOP{}) // Pending jump to end of clause
	done := blob.append(compileStmts(s, stmt.Clause.Stmts))
	blob.overwrite(jump, InstrJumpFalse{done})
	return blob
}

func compileReturnStmt(s Scope, stmt *ReturnStmt) (blob bytecode) {
	if stmt.Expr != nil {
		blob.append(compileExpr(s, stmt.Expr))
	}
	blob.write(InstrReturn{})
	return blob
}

func compileDeclarationStmt(s Scope, stmt *DeclarationStmt) bytecode {
	blob := compileExpr(s, stmt.Expr)
	symbol := s.GetVariableReference(stmt.Name.Name)
	blob.write(InstrStore{stmt.Name.Name, symbol})
	return blob
}

func compileExprStmt(s Scope, stmt *ExprStmt) bytecode {
	blob := compileExpr(s, stmt.Expr)
	blob.write(InstrPop{})
	return blob
}

func compileExpr(s Scope, expr Expr) bytecode {
	switch expr := expr.(type) {
	case *FunctionExpr:
		return compileFunctionExpr(s, expr)
	case *DispatchExpr:
		return compileDispatchExpr(s, expr)
	case *AssignExpr:
		return compileAssignExpr(s, expr)
	case *BinaryExpr:
		return compileBinaryExpr(s, expr)
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

func compileFunctionExpr(s Scope, expr *FunctionExpr) (blob bytecode) {
	local := s.GetChild(expr)
	var params []*UniqueSymbol
	for _, param := range expr.Params {
		name := param.Name.Name
		symbol := local.GetLocalVariableReference(name)
		params = append(params, symbol)
	}

	bodyBlob := bytecode{}
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
	bodyBlob.write(InstrReturn{})

	function := ObjectFunction{
		params:   params,
		bytecode: bodyBlob,
	}

	fmt.Println(bodyBlob.String())
	fmt.Println("---")

	blob.write(InstrPush{function})
	return blob
}

func compileDispatchExpr(s Scope, expr *DispatchExpr) (blob bytecode) {
	for _, arg := range expr.Args {
		blob.append(compileExpr(s, arg))
	}
	blob.append(compileExpr(s, expr.Callee))
	blob.write(InstrDispatch{args: len(expr.Args)})
	return blob
}

func compileAssignExpr(s Scope, expr *AssignExpr) bytecode {
	blob := compileExpr(s, expr.Right)
	symbol := s.GetVariableReference(expr.Left.Name)
	blob.write(InstrCopy{})
	blob.write(InstrStore{expr.Left.Name, symbol})
	return blob
}

func compileBinaryExpr(s Scope, expr *BinaryExpr) bytecode {
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

func compileSelfExpr(s Scope, expr *SelfExpr) bytecode {
	blob := bytecode{}
	blob.write(InstrLoadSelf{})
	return blob
}

func compileIdentExpr(s Scope, expr *IdentExpr) bytecode {
	blob := bytecode{}
	symbol := s.GetVariableReference(expr.Name)
	blob.write(InstrLoad{expr.Name, symbol})
	return blob
}

func compileNumberExpr(s Scope, expr *NumberExpr) (blob bytecode) {
	blob.write(InstrPush{ObjectInt{int64(expr.Val)}})
	return blob
}

func compileStringExpr(s Scope, expr *StringExpr) (blob bytecode) {
	blob.write(InstrPush{ObjectStr{expr.Val}})
	return blob
}

func compileBoolExpr(s Scope, expr *BooleanExpr) (blob bytecode) {
	blob.write(InstrPush{ObjectBool{expr.Val}})
	return blob
}
