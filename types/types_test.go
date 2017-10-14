package types

import (
	"testing"
)

var tError = Error{}
var tAny = Any{}
var tInt = Ident{"Int"}
var tBool = Ident{"Bool"}
var tOpt = Optional{tBool}
var tList = List{tInt}
var tTuple = Tuple{[]Type{tInt, tBool, tOpt, tList}}
var tFunc = Function{tTuple, tList}

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
	expectEquivalentType(t, Any{}, Any{})
	expectBool(t, tAny.Equals(tAny), true)

	expectString(t, Any{}.String(), "Any")
	expectBool(t, Any{}.IsError(), false)
	Any{}.isType()
}

func TestTypeVoid(t *testing.T) {
	expectEquivalentType(t, Void{}, Void{})
	expectNotEquivalentType(t, Void{}, Error{})
	expectNotEquivalentType(t, Void{}, tInt)
	expectBool(t, (Void{}).Equals(tAny), false)

	expectString(t, (Void{}).String(), "Void")
	expectBool(t, (Void{}).IsError(), false)
	(Void{}).isType()
}

func TestTypeFunction(t *testing.T) {
	expectEquivalentType(t, tFunc, tFunc)
	expectNotEquivalentType(t, tFunc, tList)
	expectNotEquivalentType(t, tFunc, tError)
	expectNotEquivalentType(t, tFunc, Function{tTuple, tBool})
	expectNotEquivalentType(t, tFunc, Function{Tuple{}, tList})
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
	expectNotEquivalentType(t, tTuple, Tuple{[]Type{tInt, tBool, tOpt}})
	expectNotEquivalentType(t, tTuple, Tuple{[]Type{tInt, tBool, tOpt, tOpt}})
	expectBool(t, tTuple.Equals(tAny), false)
	expectBool(t, tAny.Equals(tTuple), true)

	expectString(t, tTuple.String(), "(Int Bool Bool? [Int])")
	expectBool(t, tTuple.IsError(), false)
	tTuple.isType()
}

func TestTypeList(t *testing.T) {
	expectEquivalentType(t, List{tInt}, List{tInt})
	expectEquivalentType(t, List{tOpt}, List{tOpt})
	expectNotEquivalentType(t, List{tInt}, tInt)
	expectNotEquivalentType(t, List{tInt}, tError)
	expectNotEquivalentType(t, List{tInt}, List{tBool})
	expectBool(t, List{tInt}.Equals(tAny), false)
	expectBool(t, tAny.Equals(List{tInt}), true)

	expectString(t, List{tOpt}.String(), "[Bool?]")
	expectBool(t, tList.IsError(), false)
	tList.isType()
}

func TestTypeOptional(t *testing.T) {
	expectEquivalentType(t, Optional{tInt}, Optional{tInt})
	expectNotEquivalentType(t, Optional{tInt}, tInt)
	expectNotEquivalentType(t, Optional{tInt}, tError)
	expectNotEquivalentType(t, Optional{tInt}, Optional{tBool})
	expectBool(t, Optional{tInt}.Equals(tAny), false)
	expectBool(t, tAny.Equals(Optional{tInt}), true)

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
