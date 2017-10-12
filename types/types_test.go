package types

import (
	"testing"
)

var tError = TypeError{}
var tInt = TypeIdent{"Int"}
var tBool = TypeIdent{"Bool"}
var tOpt = TypeOptional{tBool}
var tList = TypeList{tInt}
var tTuple = TypeTuple{[]Type{tInt, tBool, tOpt, tList}}
var tFunc = TypeFunction{tTuple, tList}

func TestTypeError(t *testing.T) {
	expectEquivalentType(t, tError, tError)
	expectNotEquivalentType(t, tError, tInt)

	expectString(t, tError.String(), "ERROR")
	expectBool(t, tError.IsError(), true)
	tError.isType()
}

func TestTypeVoid(t *testing.T) {
	expectEquivalentType(t, TypeVoid{}, TypeVoid{})
	expectNotEquivalentType(t, TypeVoid{}, TypeError{})
	expectNotEquivalentType(t, TypeVoid{}, tInt)

	expectString(t, (TypeVoid{}).String(), "Void")
	expectBool(t, (TypeVoid{}).IsError(), false)
	(TypeVoid{}).isType()
}

func TestTypeFunction(t *testing.T) {
	expectEquivalentType(t, tFunc, tFunc)
	expectNotEquivalentType(t, tFunc, tList)
	expectNotEquivalentType(t, tFunc, tError)
	expectNotEquivalentType(t, tFunc, TypeFunction{tTuple, tBool})
	expectNotEquivalentType(t, tFunc, TypeFunction{TypeTuple{}, tList})

	expectString(t, tFunc.String(), "(Int Bool Bool? [Int]) => [Int]")
	expectBool(t, tFunc.IsError(), false)
	tFunc.isType()
}

func TestTypeTuple(t *testing.T) {
	expectEquivalentType(t, tTuple, tTuple)
	expectNotEquivalentType(t, tTuple, tInt)
	expectNotEquivalentType(t, tTuple, tError)
	expectNotEquivalentType(t, tTuple, TypeTuple{[]Type{tInt, tBool, tOpt}})
	expectNotEquivalentType(t, tTuple, TypeTuple{[]Type{tInt, tBool, tOpt, tOpt}})

	expectString(t, tTuple.String(), "(Int Bool Bool? [Int])")
	expectBool(t, tTuple.IsError(), false)
	tTuple.isType()
}

func TestTypeList(t *testing.T) {
	expectEquivalentType(t, TypeList{tInt}, TypeList{tInt})
	expectEquivalentType(t, TypeList{tOpt}, TypeList{tOpt})
	expectNotEquivalentType(t, TypeList{tInt}, tInt)
	expectNotEquivalentType(t, TypeList{tInt}, tError)
	expectNotEquivalentType(t, TypeList{tInt}, TypeList{tBool})

	expectString(t, TypeList{tOpt}.String(), "[Bool?]")
	expectBool(t, tList.IsError(), false)
	tList.isType()
}

func TestTypeOptional(t *testing.T) {
	expectEquivalentType(t, TypeOptional{tInt}, TypeOptional{tInt})
	expectNotEquivalentType(t, TypeOptional{tInt}, tInt)
	expectNotEquivalentType(t, TypeOptional{tInt}, tError)
	expectNotEquivalentType(t, TypeOptional{tInt}, TypeOptional{tBool})

	expectString(t, tOpt.String(), "Bool?")
	expectBool(t, tOpt.IsError(), false)
	tOpt.isType()
}

func TestTypeIdent(t *testing.T) {
	expectEquivalentType(t, tInt, tInt)
	expectNotEquivalentType(t, tInt, tError)
	expectNotEquivalentType(t, tInt, tBool)
	expectNotEquivalentType(t, tInt, tOpt)

	expectString(t, tInt.String(), "Int")
	expectString(t, tBool.String(), "Bool")
	expectBool(t, tInt.IsError(), false)
	tInt.isType()
}

func TestConcatTypes(t *testing.T) {
	expectString(t, concatTypes(nil), "")
	expectString(t, concatTypes([]Type{}), "")
	expectString(t, concatTypes([]Type{tInt, tBool, tInt}), "Int Bool Int")
}

func expectEquivalentType(t *testing.T, t1 Type, t2 Type) {
	same := t1.Equals(t2)
	commutative := t1.Equals(t2) == t2.Equals(t1)

	if commutative == false {
		if same {
			t.Errorf("%s == %s, but %s != %s", t1, t2, t2, t1)
		} else {
			t.Errorf("%s == %s, but %s != %s", t2, t1, t1, t2)
		}
	}

	if same == false {
		t.Errorf("Expected %s == %s, got %t", t1, t2, same)
	}
}

func expectNotEquivalentType(t *testing.T, t1 Type, t2 Type) {
	same := t1.Equals(t2)
	commutative := t1.Equals(t2) == t2.Equals(t1)

	if commutative == false {
		if same {
			t.Errorf("%s == %s, but %s != %s", t1, t2, t2, t1)
		} else {
			t.Errorf("%s == %s, but %s != %s", t2, t1, t1, t2)
		}
	}

	if same == true {
		t.Errorf("Expected %s != %s, got %t", t1, t2, same)
	}
}

func expectString(t *testing.T, got string, exp string) {
	if exp != got {
		t.Errorf("Expected '%s', got '%s'", exp, got)
	}
}

func expectBool(t *testing.T, got bool, exp bool) {
	if exp != got {
		t.Errorf("Expected %t, got %t", exp, got)
	}
}
