package lang

import (
	"fmt"
)

// Loc corresponds to a line & column location within source code
type Loc struct {
	Line int
	Col  int
}

func (l Loc) String() string {
	return fmt.Sprintf("(%d:%d)", l.Line, l.Col)
}

// smallerLoc takes to Loc structs and returns the Loc that occurs earlier in
// the source code
func smallerLoc(a Loc, b Loc) Loc {
	if a.Line < b.Line {
		return a
	} else if b.Line < a.Line {
		return b
	} else if a.Col < b.Col {
		return a
	}

	return b
}

// charPoint maps a character to that character's line & column within source code
type charPoint struct {
	char rune
	loc  Loc
}

// scanner holds a sequence of Char structs
type scanner struct {
	index  int
	buffer []charPoint
}

// peek returns the next character without advancing
func (cb *scanner) peek() charPoint {
	return cb.buffer[cb.index]
}

// next returns the next character and advances the buffer
func (cb *scanner) next() charPoint {
	if cb.eof() {
		return cb.buffer[cb.index]
	}

	char := cb.buffer[cb.index]
	cb.index++
	return char
}

// eof returns true if the char buffer has been exhausted
func (cb *scanner) eof() bool {
	return cb.index+1 == len(cb.buffer)
}

// scan creates a Scanner from a string of source code
func scan(source string) *scanner {
	buffer := []charPoint{}
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

		buffer = append(buffer, charPoint{char, loc})
	}

	buffer = append(buffer, charPoint{'\000', Loc{line, col}})
	return &scanner{0, buffer}
}
