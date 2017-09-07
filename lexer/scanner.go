package lexer

import (
	"fmt"
)

// Loc corresponds to a line & column location within source code
type Loc struct {
	line int
	col  int
}

func (l Loc) String() string {
	return fmt.Sprintf("(%d:%d)", l.line, l.col)
}

// Char maps a character to that character's line & column within source code
type Char struct {
	char rune
	loc  Loc
}

// Scanner holds a sequence of Char structs
type Scanner struct {
	index  int
	buffer []Char
}

// Peek returns the next character without advancing
func (cb *Scanner) Peek() Char {
	return cb.buffer[cb.index]
}

// Next returns the next character and advances the buffer
func (cb *Scanner) Next() Char {
	if cb.EOF() {
		return cb.buffer[cb.index]
	}

	char := cb.buffer[cb.index]
	cb.index++
	return char
}

// EOF returns true if the char buffer has been exhausted
func (cb *Scanner) EOF() bool {
	return cb.index+1 == len(cb.buffer)
}

// Scan creates a Scanner from a string of source code
func Scan(source string) *Scanner {
	buffer := []Char{}
	line := 1
	col := 0

	for _, char := range source {
		var loc Loc

		if char == '\n' {
			loc = Loc{line, col + 1}
			line++
			col = 0
		} else {
			col++
			loc = Loc{line, col}
		}

		buffer = append(buffer, Char{char, loc})
	}

	buffer = append(buffer, Char{'\000', Loc{line, col}})
	return &Scanner{0, buffer}
}
