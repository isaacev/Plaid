package lang

import "plaid/lang/types"

type Library struct {
	objects map[string]ObjectBuiltin
}

func (l *Library) addObject(name string, obj ObjectBuiltin) {
	l.objects[name] = obj
}

func (l *Library) toType() types.Struct {
	var fields []struct {
		Name string
		Type types.Type
	}

	for name, obj := range l.objects {
		fields = append(fields, struct {
			Name string
			Type types.Type
		}{Name: name, Type: obj.typ})
	}

	return types.Struct{fields}
}

func (l *Library) Module(name string) *ModuleNative {
	return &ModuleNative{
		name:    name,
		library: l,
	}
}

func (l *Library) toObject() ObjectStruct {
	fields := make(map[string]Object)

	for name, obj := range l.objects {
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
		objects: make(map[string]ObjectBuiltin),
	}
}
