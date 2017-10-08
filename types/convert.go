package types

import "plaid/parser"

// ConvertTypeNote transforms a TypeNote struct (used to represent a syntax
// type notation) into a Type struct (used internally to represent a type)
func ConvertTypeNote(note parser.TypeNote) Type {
	switch note := note.(type) {
	case parser.TypeNoteVoid:
		return TypeVoid{}
	case parser.TypeNoteFunction:
		return TypeFunction{ConvertTypeNote(note.Params).(TypeTuple), ConvertTypeNote(note.Ret)}
	case parser.TypeNoteTuple:
		elems := []Type{}
		for _, elem := range note.Elems {
			elems = append(elems, ConvertTypeNote(elem))
		}
		return TypeTuple{elems}
	case parser.TypeNoteList:
		return TypeList{ConvertTypeNote(note.Child)}
	case parser.TypeNoteOptional:
		return TypeOptional{ConvertTypeNote(note.Child)}
	case parser.TypeNoteIdent:
		return TypeIdent{note.Name}
	default:
		return nil
	}
}
