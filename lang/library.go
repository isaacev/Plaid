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

func (l *Library) toScope() *GlobalScope {
	s := makeGlobalScope()

	for name, sym := range l.symbols {
		obj := l.objects[sym]
		s.types[name] = obj.typ
		s.symbols[name] = sym
		s.newExport(name, obj.typ)
		s.objects[name] = obj
	}

	return s
}

func (l *Library) toType() types.Struct {
	var fields []struct {
		Name string
		Type types.Type
	}

	for name, sym := range l.symbols {
		obj := l.objects[sym]
		fields = append(fields, struct {
			Name string
			Type types.Type
		}{Name: name, Type: obj.typ})
	}

	return types.Struct{fields}
}

func (l *Library) toObject() ObjectStruct {
	fields := make(map[string]Object)

	for name, sym := range l.symbols {
		obj := l.objects[sym]
		fields[name] = obj
	}

	return ObjectStruct{fields}
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
