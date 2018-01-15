package lang

import (
	"testing"
)

func TestLoc(t *testing.T) {
	l := Loc{1, 4}

	if l.String() != "(1:4)" {
		t.Error("Expected (1:4), got ", l.String())
	}
}

func TestSmallerLoc(t *testing.T) {
	expectSmallerLoc := func(a Loc, b Loc, exp Loc) {
		greater := smallerLoc(a, b)

		if exp.String() != greater.String() {
			t.Errorf("Expected %s, got %s\n", exp, greater)
		}
	}

	expectSmallerLoc(Loc{2, 4}, Loc{2, 4}, Loc{2, 4}) // same line/col
	expectSmallerLoc(Loc{3, 1}, Loc{2, 4}, Loc{2, 4}) // a from bigger line
	expectSmallerLoc(Loc{2, 4}, Loc{3, 1}, Loc{2, 4}) // b from bigger line
	expectSmallerLoc(Loc{2, 8}, Loc{2, 4}, Loc{2, 4}) // a from bigger col
	expectSmallerLoc(Loc{2, 4}, Loc{2, 8}, Loc{2, 4}) // b from bigger col
}

func TestScanner(t *testing.T) {
	buf := scanner{0, []charPoint{
		charPoint{'a', Loc{1, 1}},
		charPoint{'b', Loc{1, 2}},
		charPoint{'c', Loc{1, 3}},
	}}

	expectChar(t, charPoint{'a', Loc{1, 1}}, buf.peek())
	expectChar(t, charPoint{'a', Loc{1, 1}}, buf.peek())
	expectIndex(t, 0, buf.index)

	expectChar(t, charPoint{'a', Loc{1, 1}}, buf.next())
	expectIndex(t, 1, buf.index)

	expectChar(t, charPoint{'b', Loc{1, 2}}, buf.next())
	expectIndex(t, 2, buf.index)

	expectChar(t, charPoint{'c', Loc{1, 3}}, buf.next())
	expectIndex(t, 2, buf.index)

	expectChar(t, charPoint{'c', Loc{1, 3}}, buf.next())
	expectIndex(t, 2, buf.index)
}

func TestScan(t *testing.T) {
	buf := scan("abc\ndef")

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

func expectNext(t *testing.T, expChar rune, expLoc Loc, buf *scanner) {
	t.Helper()
	expectChar(t, charPoint{expChar, expLoc}, buf.next())
}

func expectIndex(t *testing.T, exp int, got int) {
	t.Helper()
	if exp != got {
		t.Errorf("Expected Scanner.index %d, got %d\n", exp, got)
	}
}

func expectChar(t *testing.T, exp charPoint, got charPoint) {
	t.Helper()
	if exp.char != got.char {
		t.Errorf("Expected Char.char %c, got %c\n", exp.char, got.char)
	}

	expectLoc(t, exp.loc, got.loc)
}

func expectEOF(t *testing.T, exp bool, buf *scanner) {
	t.Helper()
	if exp != buf.eof() {
		t.Errorf("Expected Scanner#EOF() %t, got %t\n", exp, buf.eof())
	}
}

func expectLoc(t *testing.T, exp Loc, got Loc) {
	t.Helper()
	if exp.String() != got.String() {
		t.Errorf("Expected Char.loc %s, got %s\n", exp, got)
	}
}
