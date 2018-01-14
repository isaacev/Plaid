package lang

import (
	"plaid/lang/types"
)

// convertTypeNote transforms a TypeNote struct (used to represent a syntax
// type notation) into a Type struct (used internally to represent a type)
func convertTypeNote(note TypeNote) types.Type {
	switch note := note.(type) {
	case TypeNoteAny:
		return types.TypeAny{}
	case TypeNoteVoid:
		return types.TypeVoid{}
	case TypeNoteFunction:
		return types.TypeFunction{
			Params: convertTypeNote(note.Params).(types.TypeTuple),
			Ret:    convertTypeNote(note.Ret),
		}
	case TypeNoteTuple:
		elems := []types.Type{}
		for _, elem := range note.Elems {
			elems = append(elems, convertTypeNote(elem))
		}
		return types.TypeTuple{Children: elems}
	case TypeNoteList:
		return types.TypeList{Child: convertTypeNote(note.Child)}
	case TypeNoteOptional:
		return types.TypeOptional{Child: convertTypeNote(note.Child)}
	case TypeNoteIdent:
		return types.TypeIdent{Name: note.Name}
	default:
		return nil
	}
}
