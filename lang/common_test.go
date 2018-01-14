package lang

import (
	"plaid/lang/types"
	"testing"
)

func expectSame(t *testing.T, got interface{}, exp interface{}) {
	if exp != got {
		t.Errorf("Expected '%v', got '%v'", exp, got)
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

func expectEquivalentType(t *testing.T, t1 types.Type, t2 types.Type) {
	if t1 == nil {
		t.Fatalf("Expected type, got <nil>")
	}

	if t2 == nil {
		t.Fatalf("Expected type, got <nil>")
	}

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

func expectNotEquivalentType(t *testing.T, t1 types.Type, t2 types.Type) {
	if t1 == nil {
		t.Fatalf("Expected type, got <nil>")
	}

	if t2 == nil {
		t.Fatalf("Expected type, got <nil>")
	}

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

func expectNil(t *testing.T, got interface{}) {
	if got != nil {
		t.Errorf("Expected nil, got '%v'", got)
	}
}

func expectNoErrors(t *testing.T, s Scope) {
	if s.HasErrors() {
		for i, err := range s.GetErrors() {
			t.Errorf("%d '%s'", i, err)
		}

		t.Fatalf("Expected no errors, found %d", len(s.GetErrors()))
	}
}

func expectAnError(t *testing.T, err error, msg string) {
	if err == nil {
		t.Errorf("Expected an error '%s', got no errors", err)
	} else if msg != err.Error() {
		t.Errorf("Expected '%s', got '%s'", msg, err)
	}
}

func expectNthError(t *testing.T, scope Scope, n int, msg string) {
	if len(scope.GetErrors()) <= n {
		t.Fatalf("Expected at least %d errors", n+1)
	}

	expectAnError(t, scope.GetErrors()[n], msg)
}
