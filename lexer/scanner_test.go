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

	expectChar(t, Char{'a', Loc{1, 1}}, buf.Peek())
	expectChar(t, Char{'a', Loc{1, 1}}, buf.Peek())
	expectIndex(t, 0, buf.index)

	expectChar(t, Char{'a', Loc{1, 1}}, buf.Next())
	expectIndex(t, 1, buf.index)

	expectChar(t, Char{'b', Loc{1, 2}}, buf.Next())
	expectIndex(t, 2, buf.index)

	expectChar(t, Char{'c', Loc{1, 3}}, buf.Next())
	expectIndex(t, 2, buf.index)

	expectChar(t, Char{'c', Loc{1, 3}}, buf.Next())
	expectIndex(t, 2, buf.index)
}

func TestScan(t *testing.T) {
	buf := Scan("abc\ndef")

	expectNext(t, 'a', Loc{1, 1}, buf)
	expectNext(t, 'b', Loc{1, 2}, buf)
	expectNext(t, 'c', Loc{1, 3}, buf)
	expectNext(t, '\n', Loc{1, 4}, buf)
	expectNext(t, 'd', Loc{2, 1}, buf)
	expectNext(t, 'e', Loc{2, 2}, buf)
	expectNext(t, 'f', Loc{2, 3}, buf)
	expectNext(t, '\000', Loc{2, 3}, buf)
	expectNext(t, '\000', Loc{2, 3}, buf)
	expectNext(t, '\000', Loc{2, 3}, buf)
	expectEOF(t, true, buf)
}

func expectNext(t *testing.T, expChar rune, expLoc Loc, buf *CharBuffer) {
	expectChar(t, Char{expChar, expLoc}, buf.Next())
}

func expectIndex(t *testing.T, exp int, got int) {
	if exp != got {
		t.Errorf("Expected CharBuffer.index %d, got %d\n", exp, got)
	}
}

func expectChar(t *testing.T, exp Char, got Char) {
	if exp.char != got.char {
		t.Errorf("Expected Char.char %c, got %c\n", exp.char, got.char)
	}

	if exp.loc.String() != got.loc.String() {
		t.Errorf("Expected Char.loc %s, got %s\n", exp.loc, got.loc)
	}
}

func expectEOF(t *testing.T, exp bool, buf *CharBuffer) {
	if exp != buf.EOF() {
		t.Errorf("Expected CharBuffer#EOF() %t, got %t\n", exp, buf.EOF())
	}
}
