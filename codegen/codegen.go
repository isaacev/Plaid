package codegen

import (
	"fmt"
	"plaid/vm"
)

// Generate converts an IR into a Module
func Generate(ir IR) *vm.Module {
	return &vm.Module{
		Root:    genProg(ir),
		Exports: make(map[string]*vm.Export),
	}
}

func genProg(ir IR) *vm.ClosureTemplate {
	template := vm.MakeEmptyClosureTemplate(nil)

	for _, record := range ir.Scope.Local {
		template.Bytecode.Write(vm.InstrReserve{Template: recordToCellTemplate(record)})
	}

	genVoidNodes(template.Bytecode, ir.Children)
	template.Bytecode.Write(vm.InstrHalt{})
	return template
}

func genVoidNodes(bc *vm.Bytecode, nodes []IRVoidNode) {
	for _, node := range nodes {
		genVoidNode(bc, node)
	}
}

func genVoidNode(bc *vm.Bytecode, node IRVoidNode) {
	switch node := node.(type) {
	case IRCondNode:
		genCondNode(bc, node)
	case IRReturnNode:
		if node.Child != nil {
			genTypedNode(bc, node.Child)
		} else {
			bc.Write(vm.InstrNone{})
		}
		bc.Write(vm.InstrReturn{})
	case IRVoidedNode:
		genTypedNode(bc, node.Child)
		bc.Write(vm.InstrPop{})
	default:
		panic(fmt.Sprintf("cannot compile %T", node))
	}
}

func genCondNode(bc *vm.Bytecode, node IRCondNode) {
	genTypedNode(bc, node.Cond)
	jumpIP := bc.Write(vm.InstrNOP{}) // Pending jump to end of clause
	genVoidNodes(bc, node.Clause)
	doneIP := bc.NextIP()
	bc.Overwrite(jumpIP, vm.InstrJumpFalse{IP: doneIP})
}

func genTypedNode(bc *vm.Bytecode, node IRTypedNode) {
	switch node := node.(type) {
	case IRFunctionNode:
		genFunctionNode(bc, node)
	case IRDispatchNode:
		genDispatchNode(bc, node)
	case IRAssignNode:
		genAssignNode(bc, node)
	case IRBinaryNode:
		genBinaryNode(bc, node)
	case IRSelfReferenceNode:
		genSelfReferenceNode(bc, node)
	case IRReferenceNode:
		genReferenceNode(bc, node)
	case IRBuiltinReferenceNode:
		genBuiltinReferenceNode(bc, node)
	case IRIntegerLiteralNode:
		genIntegerLiteralNode(bc, node)
	case IRStringLiteralNode:
		getStringLiteralNode(bc, node)
	case IRBooleanLitearlNode:
		genBooleanLiteralNode(bc, node)
	default:
		panic(fmt.Sprintf("cannot compile %T", node))
	}
}

func genFunctionNode(bc *vm.Bytecode, node IRFunctionNode) {
	template := vm.MakeEmptyClosureTemplate(nil)

localLoop:
	for _, record := range node.Scope.Local {
		for _, param := range node.Params {
			if record == param {
				continue localLoop
			}
		}

		template.Bytecode.Write(vm.InstrReserve{Template: recordToCellTemplate(record)})
	}

	genVoidNodes(template.Bytecode, node.Body)

	for _, param := range node.Params {
		template.Parameters = append(template.Parameters, recordToCellTemplate(param))
	}

	bc.Closure.Enclose(template)
	bc.Write(vm.InstrPush{Val: template})
}

func genDispatchNode(bc *vm.Bytecode, node IRDispatchNode) {
	for _, node := range node.Args {
		genTypedNode(bc, node)
	}

	genTypedNode(bc, node.Callee)
	bc.Write(vm.InstrDispatch{NumArgs: len(node.Args)})
}

func genAssignNode(bc *vm.Bytecode, node IRAssignNode) {
	genTypedNode(bc, node.Child)
	bc.Write(vm.InstrCopy{})
	bc.Write(vm.InstrStore{Template: recordToCellTemplate(node.Record)})
}

func genBinaryNode(bc *vm.Bytecode, node IRBinaryNode) {
	genTypedNode(bc, node.Left)
	genTypedNode(bc, node.Right)

	switch node.Oper {
	case "+":
		bc.Write(vm.InstrAdd{})
	case "-":
		bc.Write(vm.InstrSub{})
	case "<":
		bc.Write(vm.InstrLT{})
	case "<=":
		bc.Write(vm.InstrLTEquals{})
	case ">":
		bc.Write(vm.InstrGT{})
	case ">=":
		bc.Write(vm.InstrGTEquals{})
	default:
		panic(fmt.Sprintf("cannot compile %T with '%s'", node, node.Oper))
	}
}

func genSelfReferenceNode(bc *vm.Bytecode, node IRSelfReferenceNode) {
	bc.Write(vm.InstrLoadSelf{})
}

func genReferenceNode(bc *vm.Bytecode, node IRReferenceNode) {
	bc.Write(vm.InstrLoad{Template: recordToCellTemplate(node.Record)})
}

func genBuiltinReferenceNode(bc *vm.Bytecode, node IRBuiltinReferenceNode) {
	obj := &vm.ObjectBuiltin{Val: node.Builtin}
	bc.Write(vm.InstrPush{Val: obj})
}

func genIntegerLiteralNode(bc *vm.Bytecode, node IRIntegerLiteralNode) {
	obj := &vm.ObjectInt{Val: node.Val}
	bc.Write(vm.InstrPush{Val: obj})
}

func getStringLiteralNode(bc *vm.Bytecode, node IRStringLiteralNode) {
	obj := &vm.ObjectStr{Val: node.Val}
	bc.Write(vm.InstrPush{Val: obj})
}

func genBooleanLiteralNode(bc *vm.Bytecode, node IRBooleanLitearlNode) {
	obj := &vm.ObjectBool{Val: node.Val}
	bc.Write(vm.InstrPush{Val: obj})
}

func recordToCellTemplate(record *VarRecord) *vm.CellTemplate {
	return &vm.CellTemplate{
		ID:   record.ID,
		Name: record.Name,
	}
}
