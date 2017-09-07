package lexer

import (
	"fmt"
)

// EOF represents the rune added at the end of a CharBuffer
const EOF = '\000'

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

// CharBuffer holds a sequence of Char structs
type CharBuffer struct {
	index  int
	buffer []Char
}

// Peek returns the next character without advancing
func (cb *CharBuffer) Peek() (Char, bool) {
	return cb.buffer[cb.index], cb.index+1 == len(cb.buffer)
}

// Next returns the next character and advances the buffer
func (cb *CharBuffer) Next() (Char, bool) {
	if cb.index+1 == len(cb.buffer) {
		return cb.buffer[cb.index], true
	}

	char := cb.buffer[cb.index]
	cb.index++
	return char, false
}

// Scan creates a CharBuffer from a string of source code
func Scan(source string) *CharBuffer {
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
	return &CharBuffer{0, buffer}
}
