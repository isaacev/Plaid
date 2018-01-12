package lang

// ConvertTypeNote transforms a TypeNote struct (used to represent a syntax
// type notation) into a Type struct (used internally to represent a type)
func ConvertTypeNote(note TypeNote) Type {
	switch note := note.(type) {
	case TypeNoteAny:
		return TypeAny{}
	case TypeNoteVoid:
		return TypeVoid{}
	case TypeNoteFunction:
		return TypeFunction{
			Params: ConvertTypeNote(note.Params).(TypeTuple),
			Ret:    ConvertTypeNote(note.Ret),
		}
	case TypeNoteTuple:
		elems := []Type{}
		for _, elem := range note.Elems {
			elems = append(elems, ConvertTypeNote(elem))
		}
		return TypeTuple{Children: elems}
	case TypeNoteList:
		return TypeList{Child: ConvertTypeNote(note.Child)}
	case TypeNoteOptional:
		return TypeOptional{Child: ConvertTypeNote(note.Child)}
	case TypeNoteIdent:
		return TypeIdent{Name: note.Name}
	default:
		return nil
	}
}
