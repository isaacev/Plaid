package check

import (
	"plaid/parser"
	"testing"
)

func TestCheckMain(t *testing.T) {
	errs := Check(parser.Program{})
	expectNoErrors(t, errs)
}

func expectNoErrors(t *testing.T, errs []error) {
	if len(errs) > 0 {
		for i, err := range errs {
			t.Errorf("%d '%s'", i, err)
		}

		t.Fatalf("Expected no errors, found %d", len(errs))
	}
}
