package lang

import "fmt"

func Run(mod *ModuleVirtual) {
	env := makeEnvironment(nil)
	mod.environment = env
	runBlob(mod, env, *mod.bytecode)
}

func loadModuleEnvironment(mod Module) {
	// If the given module has already been evaluated, do nothing.
	if mod, ok := mod.(*ModuleVirtual); ok && mod.environment == nil {
		runVirtualModule(mod)
	}
}

func runVirtualModule(mod *ModuleVirtual) {
	if mod.bytecode == nil {
		Compile(mod)
	}

	env := makeEnvironment(nil)
	mod.environment = env
	runBlob(mod, env, *mod.bytecode)
}

type environment struct {
	parent *environment
	stack  []Object
	state  map[string]Object
	self   *ObjectClosure
}

func (e *environment) pushToStack(obj Object) {
	e.stack = append(e.stack, obj)
}

func (e *environment) popFromStack() Object {
	obj := e.stack[len(e.stack)-1]
	e.stack = e.stack[:len(e.stack)-1]
	return obj
}

func (e *environment) alloc(name string) {
	e.state[name] = ObjectNone{}
}

func (e *environment) store(name string, obj Object) {
	if _, ok := e.state[name]; ok {
		e.state[name] = obj
	} else {
		e.parent.store(name, obj)
	}
}

func (e *environment) load(name string) Object {
	if obj, ok := e.state[name]; ok {
		return obj
	} else {
		if e.parent == nil {
			panic(fmt.Sprintf("cannot find variable '%s'", name))
		}
		return e.parent.load(name)
	}
}

func makeEnvironment(parent *environment) *environment {
	return &environment{
		parent: parent,
		state:  make(map[string]Object),
	}
}

func runBlob(mod *ModuleVirtual, env *environment, blob Bytecode) Object {
	var ip uint32 = 0
	instr := blob.Instructions[ip]
	for {
		if _, ok := instr.(InstrHalt); ok {
			return nil
		} else if _, ok := instr.(InstrReturn); ok {
			return env.popFromStack()
		}

		ip = runInstr(mod, ip, env, instr)
		instr = blob.Instructions[ip]
	}
}

func runInstr(mod *ModuleVirtual, ip uint32, env *environment, instr Instr) uint32 {
	switch instr := instr.(type) {
	case InstrHalt:
		return ip
	case InstrNOP:
		// do nothing
	case InstrJump:
		return uint32(instr.addr)
	case InstrJumpTrue:
		a := env.popFromStack().(*ObjectBool)
		if a.val {
			return uint32(instr.addr)
		}
	case InstrJumpFalse:
		a := env.popFromStack().(*ObjectBool)
		if a.val == false {
			return uint32(instr.addr)
		}
	case InstrPush:
		env.pushToStack(instr.Val)
	case InstrPop:
		env.popFromStack()
	case InstrCopy:
		a := env.popFromStack()
		env.pushToStack(a)
		env.pushToStack(a)
	case InstrReserve:
		env.alloc(instr.Name)
	case InstrStore:
		a := env.popFromStack()
		env.store(instr.Name, a)
	case InstrLoadMod:
		path := env.popFromStack().(*ObjectStr).val
		var alias string
		var obj Object
		for _, dep := range mod.dependencies {
			if dep.relative == path {
				if dep.module.IsNative() == false {
					runVirtualModule(dep.module.(*ModuleVirtual))
				}
				alias = dep.alias
				obj = dep.module.export()
				break
			}
		}

		if obj == nil {
			panic("could not load dependency")
		}

		env.alloc(alias)
		env.store(alias, obj)
	case InstrLoadAttr:
		a := env.popFromStack()
		env.pushToStack(a.(*ObjectStruct).Member(instr.Name))
	case InstrLoadSelf:
		env.pushToStack(env.self)
	case InstrLoad:
		a := env.load(instr.Name)
		env.pushToStack(a)
	case InstrDispatch:
		obj := env.popFromStack()
		switch fn := obj.(type) {
		case *ObjectClosure:
			child := makeEnvironment(fn.context)
			child.self = fn
			for _, sym := range fn.params {
				child.alloc(sym)
				obj := env.popFromStack()
				child.store(sym, obj)
			}
			ret := runBlob(mod, child, fn.bytecode)
			env.pushToStack(ret)
		case *ObjectBuiltin:
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
	case InstrCreateClosure:
		fn := env.popFromStack().(*ObjectFunction)
		clo := &ObjectClosure{
			context:  env,
			params:   fn.params,
			bytecode: fn.bytecode,
		}
		env.pushToStack(clo)
	case InstrAdd:
		b := env.popFromStack().(*ObjectInt)
		a := env.popFromStack().(*ObjectInt)
		sum := a.val + b.val
		env.pushToStack(&ObjectInt{sum})
	case InstrSub:
		b := env.popFromStack().(*ObjectInt)
		a := env.popFromStack().(*ObjectInt)
		sum := a.val - b.val
		env.pushToStack(&ObjectInt{sum})
	case InstrLT:
		b := env.popFromStack().(*ObjectInt)
		a := env.popFromStack().(*ObjectInt)
		test := a.val < b.val
		env.pushToStack(&ObjectBool{test})
	case InstrLTEquals:
		b := env.popFromStack().(*ObjectInt)
		a := env.popFromStack().(*ObjectInt)
		test := a.val <= b.val
		env.pushToStack(&ObjectBool{test})
	case InstrGT:
		b := env.popFromStack().(*ObjectInt)
		a := env.popFromStack().(*ObjectInt)
		test := a.val > b.val
		env.pushToStack(&ObjectBool{test})
	case InstrGTEquals:
		b := env.popFromStack().(*ObjectInt)
		a := env.popFromStack().(*ObjectInt)
		test := a.val >= b.val
		env.pushToStack(&ObjectBool{test})
	default:
		panic(fmt.Sprintf("cannot interpret %T instructions", instr))
	}

	return ip + 1
}
