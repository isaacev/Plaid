package lang

import (
	"plaid/lang/types"
)

// convertTypeNote transforms a TypeNote struct (used to represent a syntax
// type notation) into a Type struct (used internally to represent a type)
func convertTypeNote(note TypeNote) types.Type {
	switch note := note.(type) {
	case TypeNoteAny:
		return types.Any{}
	case TypeNoteVoid:
		return types.Void{}
	case TypeNoteFunction:
		return types.Function{
			Params: convertTypeNote(note.Params).(types.Tuple),
			Ret:    convertTypeNote(note.Ret),
		}
	case TypeNoteTuple:
		elems := []types.Type{}
		for _, elem := range note.Elems {
			elems = append(elems, convertTypeNote(elem))
		}
		return types.Tuple{Children: elems}
	case TypeNoteList:
		return types.List{Child: convertTypeNote(note.Child)}
	case TypeNoteOptional:
		return types.Optional{Child: convertTypeNote(note.Child)}
	case TypeNoteIdent:
		return types.Ident{Name: note.Name}
	default:
		return nil
	}
}
