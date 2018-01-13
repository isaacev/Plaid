package lang

// convertTypeNote transforms a TypeNote struct (used to represent a syntax
// type notation) into a Type struct (used internally to represent a type)
func convertTypeNote(note TypeNote) Type {
	switch note := note.(type) {
	case TypeNoteAny:
		return TypeAny{}
	case TypeNoteVoid:
		return TypeVoid{}
	case TypeNoteFunction:
		return TypeFunction{
			Params: convertTypeNote(note.Params).(TypeTuple),
			Ret:    convertTypeNote(note.Ret),
		}
	case TypeNoteTuple:
		elems := []Type{}
		for _, elem := range note.Elems {
			elems = append(elems, convertTypeNote(elem))
		}
		return TypeTuple{Children: elems}
	case TypeNoteList:
		return TypeList{Child: convertTypeNote(note.Child)}
	case TypeNoteOptional:
		return TypeOptional{Child: convertTypeNote(note.Child)}
	case TypeNoteIdent:
		return TypeIdent{Name: note.Name}
	default:
		return nil
	}
}
