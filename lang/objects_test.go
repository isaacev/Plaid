package lang

import "testing"

func TestObjectNone(t *testing.T) {
	obj := &ObjectNone{}
	obj.isObject()
	expectString(t, obj.String(), "<none>")
}

func TestObjectInt(t *testing.T) {
	obj := &ObjectInt{val: 123}
	obj.isObject()
	expectString(t, obj.String(), "123")
}

func TestObjectStr(t *testing.T) {
	obj := &ObjectStr{val: "abc"}
	obj.isObject()
	expectString(t, obj.String(), `"abc"`)
}

func TestObjectBool(t *testing.T) {
	obj := &ObjectBool{val: true}
	obj.isObject()
	expectString(t, obj.String(), "true")
}

func TestObjectBuiltin(t *testing.T) {
	obj := &ObjectBuiltin{}
	obj.isObject()
	expectString(t, obj.String(), "<builtin>")
}

func TestObjectFunction(t *testing.T) {
	obj := &ObjectFunction{}
	obj.isObject()
	expectString(t, obj.String(), "<function>")
}

func TestObjectClosure(t *testing.T) {
	obj := &ObjectClosure{}
	obj.isObject()
	expectString(t, obj.String(), "<closure>")
}
