package codegen

import (
	"fmt"
	"plaid/vm"
)

// Gen converts an AST to a Bytecode sequence
func Gen(ir IR) *vm.Bytecode {
	bc := &vm.Bytecode{}

	for _, record := range ir.Scope.Local {
		bc.Write(vm.InstrReserve{Template: recordToCellTemplate(record)})
	}

	genVoidNodes(bc, ir.Children)

	bc.Write(vm.InstrHalt{})
	return bc
}

func genVoidNodes(bc *vm.Bytecode, nodes []IRVoidNode) {
	for _, node := range nodes {
		genVoidNode(bc, node)
	}
}

func genVoidNode(bc *vm.Bytecode, node IRVoidNode) {
	switch node := node.(type) {
	case IRPrintNode:
		genTypedNode(bc, node.Child)
		bc.Write(vm.InstrPrint{})
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
	case IRReferenceNode:
		genReferenceNode(bc, node)
	case IRIntegerLiteralNode:
		genIntegerLiteralNode(bc, node)
	case IRStringLiteralNode:
		getStringLiteralNode(bc, node)
	default:
		panic(fmt.Sprintf("cannot compile %T", node))
	}
}

func genFunctionNode(bc *vm.Bytecode, node IRFunctionNode) {
	fnbc := &vm.Bytecode{}

localLoop:
	for _, record := range node.Scope.Local {
		for _, param := range node.Params {
			if record == param {
				continue localLoop
			}
		}

		fnbc.Write(vm.InstrReserve{Template: recordToCellTemplate(record)})
	}

	genVoidNodes(fnbc, node.Body)

	var params []*vm.CellTemplate
	for _, param := range node.Params {
		params = append(params, recordToCellTemplate(param))
	}

	obj := &vm.ClosureTemplate{
		Parameters: params,
		Bytecode:   fnbc,
	}
	bc.Write(vm.InstrPush{Val: obj})
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
	default:
		panic(fmt.Sprintf("cannot compile %T with '%s'", node, node.Oper))
	}
}

func genReferenceNode(bc *vm.Bytecode, node IRReferenceNode) {
	bc.Write(vm.InstrLoad{Template: recordToCellTemplate(node.Record)})
}

func genIntegerLiteralNode(bc *vm.Bytecode, node IRIntegerLiteralNode) {
	obj := &vm.ObjectInt{Val: node.Val}
	bc.Write(vm.InstrPush{Val: obj})
}

func getStringLiteralNode(bc *vm.Bytecode, node IRStringLiteralNode) {
	obj := &vm.ObjectStr{Val: node.Val}
	bc.Write(vm.InstrPush{Val: obj})
}

func recordToCellTemplate(record *VarRecord) *vm.CellTemplate {
	return &vm.CellTemplate{
		ID:   record.ID,
		Name: record.Name,
	}
}
