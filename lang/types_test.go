package lang

import (
	"testing"
)

var tError = TypeError{}
var tAny = TypeAny{}
var tInt = TypeIdent{"Int"}
var tBool = TypeIdent{"Bool"}
var tOpt = TypeOptional{tBool}
var tList = TypeList{tInt}
var tTuple = TypeTuple{[]Type{tInt, tBool, tOpt, tList}}
var tFunc = TypeFunction{tTuple, tList}

func TestTypeError(t *testing.T) {
	expectEquivalentType(t, tError, tError)
	expectNotEquivalentType(t, tError, tInt)
	expectBool(t, tError.Equals(tAny), false)
	expectBool(t, tAny.Equals(tError), false)

	expectString(t, tError.String(), "ERROR")
	expectBool(t, tError.IsError(), true)
	tError.isType()
}

func TestTypeAny(t *testing.T) {
	expectEquivalentType(t, TypeAny{}, TypeAny{})
	expectBool(t, tAny.Equals(tAny), true)
	expectBool(t, tAny.Equals(TypeVoid{}), false)

	expectString(t, TypeAny{}.String(), "Any")
	expectBool(t, TypeAny{}.IsError(), false)
	TypeAny{}.isType()
}

func TestTypeVoid(t *testing.T) {
	expectEquivalentType(t, TypeVoid{}, TypeVoid{})
	expectNotEquivalentType(t, TypeVoid{}, TypeError{})
	expectNotEquivalentType(t, TypeVoid{}, tInt)
	expectBool(t, (TypeVoid{}).Equals(tAny), false)

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
	expectBool(t, tFunc.Equals(tAny), false)
	expectBool(t, tAny.Equals(tFunc), true)

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
	expectBool(t, tTuple.Equals(tAny), false)
	expectBool(t, tAny.Equals(tTuple), true)

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
	expectBool(t, TypeList{tInt}.Equals(tAny), false)
	expectBool(t, tAny.Equals(TypeList{tInt}), true)

	expectString(t, TypeList{tOpt}.String(), "[Bool?]")
	expectBool(t, tList.IsError(), false)
	tList.isType()
}

func TestTypeOptional(t *testing.T) {
	expectEquivalentType(t, TypeOptional{tInt}, TypeOptional{tInt})
	expectNotEquivalentType(t, TypeOptional{tInt}, tInt)
	expectNotEquivalentType(t, TypeOptional{tInt}, tError)
	expectNotEquivalentType(t, TypeOptional{tInt}, TypeOptional{tBool})
	expectBool(t, TypeOptional{tInt}.Equals(tAny), false)
	expectBool(t, tAny.Equals(TypeOptional{tInt}), true)

	expectString(t, tOpt.String(), "Bool?")
	expectBool(t, tOpt.IsError(), false)
	tOpt.isType()
}

func TestTypeIdent(t *testing.T) {
	expectEquivalentType(t, tInt, tInt)
	expectNotEquivalentType(t, tInt, tError)
	expectNotEquivalentType(t, tInt, tBool)
	expectNotEquivalentType(t, tInt, tOpt)
	expectBool(t, tInt.Equals(tAny), false)
	expectBool(t, tAny.Equals(tInt), true)

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
