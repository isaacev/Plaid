package lang

import "fmt"

func Run(btc Bytecode) {
	env := makeEnvironment(nil)
	runBlob(env, btc)
}

type environment struct {
	parent *environment
	stack  []Object
	state  map[*UniqueSymbol]Object
}

func (e *environment) pushToStack(obj Object) {
	e.stack = append(e.stack, obj)
}

func (e *environment) popFromStack() Object {
	obj := e.stack[len(e.stack)-1]
	e.stack = e.stack[:len(e.stack)-1]
	return obj
}

func (e *environment) alloc(sym *UniqueSymbol) {
	e.state[sym] = ObjectNone{}
}

func (e *environment) store(sym *UniqueSymbol, obj Object) {
	if _, ok := e.state[sym]; ok {
		e.state[sym] = obj
	} else {
		e.parent.store(sym, obj)
	}
}

func (e *environment) load(sym *UniqueSymbol) Object {
	if obj, ok := e.state[sym]; ok {
		return obj
	} else {
		return e.parent.load(sym)
	}
}

func makeEnvironment(parent *environment) *environment {
	return &environment{
		parent: parent,
		state:  make(map[*UniqueSymbol]Object),
	}
}

func runBlob(env *environment, blob Bytecode) Object {
	fmt.Println("---")
	var ip uint32 = 0
	instr := blob.Instructions[ip]
	for {
		fmt.Printf("%s %s\n", Address(ip), instr)
		if _, ok := instr.(InstrHalt); ok {
			fmt.Println("---")
			return nil
		} else if _, ok := instr.(InstrReturn); ok {
			fmt.Println("---")
			return env.popFromStack()
		}

		ip = runInstr(ip, env, instr)
		instr = blob.Instructions[ip]
	}
}

func runInstr(ip uint32, env *environment, instr Instr) uint32 {
	switch instr := instr.(type) {
	case InstrHalt:
		return ip
	case InstrNOP:
		// do nothing
	case InstrPush:
		env.pushToStack(instr.Val)
	case InstrPop:
		env.popFromStack()
	case InstrReserve:
		env.alloc(instr.Symbol)
	case InstrStore:
		a := env.popFromStack()
		env.store(instr.Symbol, a)
	case InstrLoadAttr:
		a := env.popFromStack()
		env.pushToStack(a.(ObjectStruct).Member(instr.Name))
	case InstrLoad:
		a := env.load(instr.Symbol)
		env.pushToStack(a)
	case InstrDispatch:
		obj := env.popFromStack()
		switch fn := obj.(type) {
		case ObjectFunction:
			child := makeEnvironment(env)
			for _, sym := range fn.params {
				child.alloc(sym)
				obj := env.popFromStack()
				child.store(sym, obj)
			}
			ret := runBlob(child, fn.bytecode)
			env.pushToStack(ret)
		case ObjectBuiltin:
			var args []Object
			for i := 0; i < instr.args; i++ {
				args = append(args, env.popFromStack())
			}
			if ret, err := fn.val(args); err != nil {
				panic(err)
			} else {
				env.pushToStack(ret)
			}
		default:
			panic(fmt.Sprintf("cannot call %T", obj))
		}
	case InstrAdd:
		b := env.popFromStack().(ObjectInt)
		a := env.popFromStack().(ObjectInt)
		sum := a.val + b.val
		env.pushToStack(&ObjectInt{sum})
	case InstrSub:
		b := env.popFromStack().(ObjectInt)
		a := env.popFromStack().(ObjectInt)
		sum := a.val - b.val
		env.pushToStack(&ObjectInt{sum})
	case InstrPrint:
		a := env.popFromStack()
		fmt.Println(a.String())
	default:
		panic(fmt.Sprintf("cannot interpret %T instructions", instr))
	}

	return ip + 1
}
