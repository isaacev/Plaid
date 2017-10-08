package types

import (
	"plaid/lexer"
	"plaid/parser"
	"testing"
)

var nop lexer.Token

func TestConvertTypeNote(t *testing.T) {
	var note parser.TypeNote
	var blob parser.TypeNote = parser.TypeNoteIdent{Tok: nop, Name: "Blob"}
	var blub parser.TypeNote = parser.TypeNoteIdent{Tok: nop, Name: "Blub"}
	var blah parser.TypeNote = parser.TypeNoteIdent{Tok: nop, Name: "Blah"}

	note = parser.TypeNoteVoid{}
	expectConversion(t, note, "Void")

	note = parser.TypeNoteFunction{
		Params: parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{}},
		Ret:    parser.TypeNoteVoid{},
	}
	expectConversion(t, note, "() => Void")

	note = parser.TypeNoteFunction{
		Params: parser.TypeNoteTuple{Tok: nop, Elems: []parser.TypeNote{
			blob,
			blub,
		}},
		Ret: blah,
	}
	expectConversion(t, note, "(Blob Blub) => Blah")

	note = parser.TypeNoteList{Tok: nop, Child: blob}
	expectConversion(t, note, "[Blob]")

	note = parser.TypeNoteOptional{Tok: nop, Child: blah}
	expectConversion(t, note, "Blah?")

	note = nil
	got := ConvertTypeNote(note)
	if got != nil {
		t.Errorf("Expected '%v', got '%v'", nil, got)
	}
}

func expectConversion(t *testing.T, note parser.TypeNote, exp string) {
	got := ConvertTypeNote(note)
	if got.String() != exp {
		t.Errorf("Expected '%s', got '%v'", exp, got)
	}
}
