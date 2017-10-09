package vm

import (
	"fmt"
)

// Register holds a value and is associated with a variable
type Register struct {
	id  uint8
	val Object
}

func (r *Register) setValue(newVal Object) {
	r.val = newVal
}

func (r *Register) String() string {
	return fmt.Sprintf("r%d", r.id)
}

// Run does stuff
func Run(bc *Bytecode) {
	env := makeEnv(nil)
	runWithEnv(env, bc)
}

func runWithEnv(env *Env, bc *Bytecode) {
	var ip uint32
	instr := bc.Instrs[ip]
	for {
		if _, ok := instr.(InstrHalt); ok {
			return
		}

		if _, ok := instr.(InstrReturn); ok {
			return
		}

		ip = runInstr(ip, env, instr)
		instr = bc.Instrs[ip]
	}
}

func runInstr(ip uint32, env *Env, instr Instr) uint32 {
	switch instr := instr.(type) {
	case InstrHalt:
		return ip
	case InstrNOP:
		// do nothing
	case InstrJump:
		return instr.IP
	case InstrJumpTrue:
		if obj, ok := env.pop().(*ObjectBool); ok {
			if obj.Val {
				return instr.IP
			}
		} else {
			panic(fmt.Sprintf("expected boolean, got %T", obj))
		}
	case InstrJumpFalse:
		if obj, ok := env.pop().(*ObjectBool); ok {
			if obj.Val == false {
				return instr.IP
			}
		} else {
			panic(fmt.Sprintf("expected boolean, got %T", obj))
		}
	case InstrPush:
		if obj, ok := instr.Val.(*ClosureTemplate); ok {
			env.push(buildClosureFromTemplate(env, obj))
		} else {
			env.push(instr.Val)
		}
	case InstrPop:
		env.pop()
	case InstrCopy:
		obj := env.pop()
		env.push(obj)
		env.push(obj)
	case InstrReserve:
		env.reserve(instr.Template)
	case InstrStore:
		obj := env.pop()
		env.store(instr.Template, obj)
	case InstrLoad:
		obj := env.load(instr.Template)
		env.push(obj)
	case InstrDispatch:
		runInstrDispatch(env, instr)
	case InstrNone:
		env.push(&ObjectNone{})
	case InstrAdd:
		runInstrAdd(env, instr)
	case InstrSub:
		runInstrSub(env, instr)
	case InstrPrint:
		runInstrPrint(env, instr)
	}

	return ip + 1
}

func runInstrDispatch(env *Env, instr InstrDispatch) {
	popped := env.pop()
	if callee, ok := popped.(*Closure); ok {
		subEnv := makeEnv(callee.Env)

		for i := instr.NumArgs - 1; i >= 0; i-- {
			arg := env.pop()
			subEnv.reserve(callee.Parameters[i])
			subEnv.store(callee.Parameters[i], arg)
		}

		runWithEnv(subEnv, callee.Bytecode)
		output := subEnv.pop()
		env.push(output)
	} else if obj, ok := popped.(*ObjectBuiltin); ok {
		var args []Object
		for i := 0; i < instr.NumArgs; i++ {
			arg := env.pop()
			args = append(args, arg)
		}

		output, err := obj.Val.Func(args)
		if err != nil {
			panic(err.Error())
		} else {
			env.push(output)
		}
	}
}

func runInstrAdd(env *Env, instr InstrAdd) {
	b := env.pop().(*ObjectInt)
	a := env.pop().(*ObjectInt)
	sum := a.Val + b.Val
	env.push(&ObjectInt{sum})
}

func runInstrSub(env *Env, instr InstrSub) {
	b := env.pop().(*ObjectInt)
	a := env.pop().(*ObjectInt)
	sum := a.Val - b.Val
	env.push(&ObjectInt{sum})
}

func runInstrPrint(env *Env, instr InstrPrint) {
	obj := env.pop()
	fmt.Println(obj)
}

func buildClosureFromTemplate(env *Env, obj *ClosureTemplate) *Closure {
	return &Closure{
		Env:        env,
		Parameters: obj.Parameters,
		Bytecode:   obj.Bytecode,
	}
}
