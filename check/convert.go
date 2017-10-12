package check

import (
	"plaid/parser"
	"plaid/types"
)

// ConvertTypeNote transforms a TypeNote struct (used to represent a syntax
// type notation) into a Type struct (used internally to represent a type)
func ConvertTypeNote(note parser.TypeNote) types.Type {
	switch note := note.(type) {
	case parser.TypeNoteVoid:
		return types.TypeVoid{}
	case parser.TypeNoteFunction:
		return types.TypeFunction{
			Params: ConvertTypeNote(note.Params).(types.TypeTuple),
			Ret:    ConvertTypeNote(note.Ret),
		}
	case parser.TypeNoteTuple:
		elems := []types.Type{}
		for _, elem := range note.Elems {
			elems = append(elems, ConvertTypeNote(elem))
		}
		return types.TypeTuple{Children: elems}
	case parser.TypeNoteList:
		return types.TypeList{Child: ConvertTypeNote(note.Child)}
	case parser.TypeNoteOptional:
		return types.TypeOptional{Child: ConvertTypeNote(note.Child)}
	case parser.TypeNoteIdent:
		return types.TypeIdent{Name: note.Name}
	default:
		return nil
	}
}
