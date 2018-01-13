package lang

import (
	"testing"
)

func TestConvertTypeNote(t *testing.T) {
	var note TypeNote
	var blob TypeNote = TypeNoteIdent{Tok: nop, Name: "Blob"}
	var blub TypeNote = TypeNoteIdent{Tok: nop, Name: "Blub"}
	var blah TypeNote = TypeNoteIdent{Tok: nop, Name: "Blah"}

	note = TypeNoteVoid{}
	expectConversion(t, note, "Void")

	note = TypeNoteAny{}
	expectConversion(t, note, "Any")

	note = TypeNoteFunction{
		Params: TypeNoteTuple{Tok: nop, Elems: []TypeNote{}},
		Ret:    TypeNoteVoid{},
	}
	expectConversion(t, note, "() => Void")

	note = TypeNoteFunction{
		Params: TypeNoteTuple{Tok: nop, Elems: []TypeNote{
			blob,
			blub,
		}},
		Ret: blah,
	}
	expectConversion(t, note, "(Blob Blub) => Blah")

	note = TypeNoteList{Tok: nop, Child: blob}
	expectConversion(t, note, "[Blob]")

	note = TypeNoteOptional{Tok: nop, Child: blah}
	expectConversion(t, note, "Blah?")

	note = nil
	got := convertTypeNote(note)
	if got != nil {
		t.Errorf("Expected '%v', got '%v'", nil, got)
	}
}

func expectConversion(t *testing.T, note TypeNote, exp string) {
	got := convertTypeNote(note)
	if got.String() != exp {
		t.Errorf("Expected '%s', got '%v'", exp, got)
	}
}
