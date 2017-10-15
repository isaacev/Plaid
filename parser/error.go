package parser

import (
	"fmt"
	"plaid/lexer"
)

// SyntaxError combines a source code location with the resulting error message
type SyntaxError struct {
	Filepath string
	Location lexer.Loc
	Message  string
}

func (err SyntaxError) Error() string {
	return fmt.Sprintf("%s%s %s", err.Filepath, err.Location, err.Message)
}
