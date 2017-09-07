package lexer

import (
	"testing"
)

func TestLoc(t *testing.T) {
	l := Loc{1, 4}

	if l.String() != "(1:4)" {
		t.Error("Expected (1:4), got ", l.String())
	}
}

func TestCharBuffer(t *testing.T) {
	buf := CharBuffer{0, []Char{
		Char{'a', Loc{1, 1}},
		Char{'b', Loc{1, 2}},
		Char{'c', Loc{1, 3}},
	}}

	var char Char
	var eof bool

	char, eof = buf.Peek()
	expectChar(t, Char{'a', Loc{1, 1}}, false, char, eof)

	char, eof = buf.Peek()
	expectChar(t, Char{'a', Loc{1, 1}}, false, char, eof)
	expectIndex(t, 0, buf.index)

	char, eof = buf.Next()
	expectChar(t, Char{'a', Loc{1, 1}}, false, char, eof)
	expectIndex(t, 1, buf.index)

	char, eof = buf.Next()
	expectChar(t, Char{'b', Loc{1, 2}}, false, char, eof)
	expectIndex(t, 2, buf.index)

	char, eof = buf.Next()
	expectChar(t, Char{'c', Loc{1, 3}}, true, char, eof)
	expectIndex(t, 2, buf.index)

	char, eof = buf.Next()
	expectChar(t, Char{'c', Loc{1, 3}}, true, char, eof)
	expectIndex(t, 2, buf.index)
}

func TestScan(t *testing.T) {
	buf := Scan("abc\ndef")

	expectNext(t, 'a', Loc{1, 1}, false, buf)
	expectNext(t, 'b', Loc{1, 2}, false, buf)
	expectNext(t, 'c', Loc{1, 3}, false, buf)
	expectNext(t, '\n', Loc{1, 4}, false, buf)
	expectNext(t, 'd', Loc{2, 1}, false, buf)
	expectNext(t, 'e', Loc{2, 2}, false, buf)
	expectNext(t, 'f', Loc{2, 3}, false, buf)
	expectNext(t, EOF, Loc{2, 3}, true, buf)
	expectNext(t, EOF, Loc{2, 3}, true, buf)
	expectNext(t, EOF, Loc{2, 3}, true, buf)
}

func expectNext(t *testing.T, expChar rune, expLoc Loc, expEOF bool, buf *CharBuffer) {
	got, gotEOF := buf.Next()
	expectChar(t, Char{expChar, expLoc}, expEOF, got, gotEOF)
}

func expectIndex(t *testing.T, exp int, got int) {
	if exp != got {
		t.Errorf("Expected CharBuffer.index %d, got %d\n", exp, got)
	}
}

func expectChar(t *testing.T, exp Char, expEOF bool, got Char, gotEOF bool) {
	if exp.char != got.char {
		t.Errorf("Expected Char.char %c, got %c\n", exp.char, got.char)
	}

	if exp.loc.String() != got.loc.String() {
		t.Errorf("Expected Char.loc %s, got %s\n", exp.loc, got.loc)
	}

	if expEOF != gotEOF {
		t.Errorf("Expected Char.eof %t, got %t\n", expEOF, gotEOF)
	}
}
