package lang

import "plaid/lang/types"

type Library struct {
	symbols map[string]*UniqueSymbol
	objects map[*UniqueSymbol]ObjectBuiltin
}

func (l *Library) addObject(name string, obj ObjectBuiltin) {
	sym := &UniqueSymbol{name}
	l.symbols[name] = sym
	l.objects[sym] = obj
}

func (l *Library) Function(name string, typ types.Function, fn func(args []Object) (Object, error)) {
	l.addObject(name, ObjectBuiltin{
		typ: typ,
		val: fn,
	})
}

func MakeLibrary(name string) *Library {
	return &Library{
		symbols: make(map[string]*UniqueSymbol),
		objects: make(map[*UniqueSymbol]ObjectBuiltin),
	}
}
