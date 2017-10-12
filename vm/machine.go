package vm

import (
	"fmt"
)

func (r *Register) String() string {
	return fmt.Sprintf("r%d", r.Template.ID)
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
		env.reserve(instr.Register)
	case InstrStore:
		obj := env.pop()
		env.store(instr.Register, obj)
	case InstrLoadSelf:
		obj := env.Self
		env.push(obj)
	case InstrLoad:
		obj := env.load(instr.Register)
		env.push(obj)
	case InstrDispatch:
		runInstrDispatch(env, instr)
	case InstrNone:
		env.push(&ObjectNone{})
	case InstrAdd:
		runInstrAdd(env, instr)
	case InstrSub:
		runInstrSub(env, instr)
	case InstrLT:
		runInstrLT(env, instr)
	case InstrLTEquals:
		runInstrLTEquals(env, instr)
	case InstrGT:
		runInstrGT(env, instr)
	case InstrGTEquals:
		runInstrGTEquals(env, instr)
	}

	return ip + 1
}

func runInstrDispatch(env *Env, instr InstrDispatch) {
	popped := env.pop()
	if callee, ok := popped.(*Closure); ok {
		subEnv := makeEnv(callee.Env)
		subEnv.Self = callee

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

func runInstrLT(env *Env, instr InstrLT) {
	b := env.pop().(*ObjectInt)
	a := env.pop().(*ObjectInt)
	test := a.Val < b.Val
	env.push(&ObjectBool{test})
}

func runInstrLTEquals(env *Env, instr InstrLTEquals) {
	b := env.pop().(*ObjectInt)
	a := env.pop().(*ObjectInt)
	test := a.Val <= b.Val
	env.push(&ObjectBool{test})
}

func runInstrGT(env *Env, instr InstrGT) {
	b := env.pop().(*ObjectInt)
	a := env.pop().(*ObjectInt)
	test := a.Val > b.Val
	env.push(&ObjectBool{test})
}

func runInstrGTEquals(env *Env, instr InstrGTEquals) {
	b := env.pop().(*ObjectInt)
	a := env.pop().(*ObjectInt)
	test := a.Val >= b.Val
	env.push(&ObjectBool{test})
}

func buildClosureFromTemplate(env *Env, obj *ClosureTemplate) *Closure {
	return &Closure{
		Env:        env,
		Parameters: obj.Parameters,
		Bytecode:   obj.Bytecode,
	}
}
