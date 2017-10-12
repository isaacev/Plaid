package vm

import "fmt"

// Env holds variable references available in the current context
type Env struct {
	Parent    *Env
	Imports   map[int]*Register
	Stack     []Object
	Registers map[int]*Register
	Self      *Closure
}

// Import applies the given module's exports to the environment
func (env *Env) Import(module *Module) {
	for _, export := range module.Exports {
		env.Imports[export.Register.ID] = &Register{
			Contents: export.Object,
			Template: export.Register,
		}
	}
}

func (env *Env) hasParent() bool {
	return (env.Parent != nil)
}

func (env *Env) push(obj Object) {
	env.Stack = append(env.Stack, obj)
}

func (env *Env) pop() Object {
	obj := env.Stack[len(env.Stack)-1]
	env.Stack = env.Stack[:len(env.Stack)-1]
	return obj
}

func (env *Env) reserve(reg *RegisterTemplate) {
	env.Registers[reg.ID] = &Register{
		Contents: nil,
		Template: reg,
	}
}

func (env *Env) store(reg *RegisterTemplate, obj Object) {
	local, ok := env.Registers[reg.ID]
	if ok {
		local.Contents = obj
	} else if env.hasParent() {
		env.Parent.store(reg, obj)
	} else {
		panic(fmt.Sprintf("%s not local or remote", reg.Name))
	}
}

func (env *Env) load(reg *RegisterTemplate) Object {
	local, ok := env.Registers[reg.ID]
	if ok {
		return local.Contents
	} else if env.hasParent() {
		return env.Parent.load(reg)
	} else {
		global, ok := env.Imports[reg.ID]
		if ok {
			return global.Contents
		}

		fmt.Println(env.Registers, reg.ID)
		panic(fmt.Sprintf("%s not in scope", reg.Name))
	}
}

func makeEnv(parent *Env) *Env {
	return &Env{
		Parent:    parent,
		Imports:   make(map[int]*Register),
		Stack:     []Object{},
		Registers: make(map[int]*Register),
		Self:      nil,
	}
}
