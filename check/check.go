package check

import (
	"plaid/parser"
)

// Check takes an existing abstract syntax tree and performs type checks and
// other correctness checks. It returns a list of any errors that were
// discovered inside the AST
func Check(prog parser.Program) []error {
	return nil
}
