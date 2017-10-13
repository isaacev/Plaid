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
		return types.Void{}
	case parser.TypeNoteFunction:
		return types.Function{
			Params: ConvertTypeNote(note.Params).(types.Tuple),
			Ret:    ConvertTypeNote(note.Ret),
		}
	case parser.TypeNoteTuple:
		elems := []types.Type{}
		for _, elem := range note.Elems {
			elems = append(elems, ConvertTypeNote(elem))
		}
		return types.Tuple{Children: elems}
	case parser.TypeNoteList:
		return types.List{Child: ConvertTypeNote(note.Child)}
	case parser.TypeNoteOptional:
		return types.Optional{Child: ConvertTypeNote(note.Child)}
	case parser.TypeNoteIdent:
		return types.Ident{Name: note.Name}
	default:
		return nil
	}
}
