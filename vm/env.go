package vm

import "fmt"

// Env holds variable references available in the current context
type Env struct {
	Parent *Env
	Stack  []Object
	Cells  map[uint]*Cell
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

func (env *Env) reserve(template *CellTemplate) {
	env.Cells[template.ID] = &Cell{
		ID:  template.ID,
		Ref: nil,
	}
}

func (env *Env) store(template *CellTemplate, obj Object) {
	local, ok := env.Cells[template.ID]
	if ok {
		local.Ref = obj
	} else if env.hasParent() {
		env.Parent.store(template, obj)
	} else {
		panic(fmt.Sprintf("%s not local or remote", template.Name))
	}
}

func (env *Env) load(template *CellTemplate) Object {
	local, ok := env.Cells[template.ID]
	if ok {
		return local.Ref
	} else if env.hasParent() {
		return env.Parent.load(template)
	} else {
		panic(fmt.Sprintf("%s not local or remote", template.Name))
	}
}

func makeEnv(parent *Env) *Env {
	return &Env{
		Parent: parent,
		Stack:  []Object{},
		Cells:  make(map[uint]*Cell),
	}
}
