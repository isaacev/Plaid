package parser

import (
	"fmt"
	"plaid/lexer"
)

// SyntaxError combines a source code location with the resulting error message
type SyntaxError struct {
	loc lexer.Loc
	msg string
}

func (se SyntaxError) Error() string {
	return fmt.Sprintf("%s %s", se.loc, se.msg)
}

func makeSyntaxError(tok lexer.Token, msg string, deference bool) error {
	if tok.Type == lexer.Error && deference {
		msg = tok.Lexeme
	}

	return SyntaxError{tok.Loc, msg}
}
