package lang

import (
	"fmt"
)

// Type describes all type structs
type Type interface {
	Equals(other Type) bool
	IsError() bool
	String() string
	isType()
}

// TypeError signals that whatever expression was supposed to produce this type
// had a semantic error that made proper evaluation impossible
type TypeError struct{}

// Equals returns false for every other type
func (t TypeError) Equals(other Type) bool {
	if _, ok := other.(TypeError); ok {
		return true
	}

	return false
}

// IsError returns true because TypeError is an error
func (t TypeError) IsError() bool  { return true }
func (t TypeError) String() string { return "ERROR" }
func (t TypeError) isType()        {}

// TypeAny can represent all types
type TypeAny struct{}

// Equals returns true for every other type
func (t TypeAny) Equals(other Type) bool {
	if other.IsError() {
		return false
	} else if _, ok := other.(TypeVoid); ok {
		return false
	}

	return true
}

// IsError returns false because this is a properly resolved type
func (t TypeAny) IsError() bool  { return false }
func (t TypeAny) String() string { return "Any" }
func (t TypeAny) isType()        {}

// TypeVoid descirbes the return type of a function that returns no value
type TypeVoid struct{}

// Equals returns true if another type has an identical structure and identical names
func (t TypeVoid) Equals(other Type) bool {
	if _, ok := other.(TypeVoid); ok {
		return true
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t TypeVoid) IsError() bool  { return false }
func (t TypeVoid) String() string { return "Void" }
func (t TypeVoid) isType()        {}

// TypeFunction describes mappings of 0+ parameter types to a return type
type TypeFunction struct {
	Params TypeTuple
	Ret    Type
}

// Equals returns true if another type has an identical structure and identical names
func (t TypeFunction) Equals(other Type) bool {
	if t2, ok := other.(TypeFunction); ok {
		return t.Params.Equals(t2.Params) && t.Ret.Equals(t2.Ret)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t TypeFunction) IsError() bool  { return false }
func (t TypeFunction) String() string { return fmt.Sprintf("%s => %s", t.Params, t.Ret) }
func (t TypeFunction) isType()        {}

// TypeTuple describes a group of types
type TypeTuple struct {
	Children []Type
}

// Equals returns true if another type has an identical structure and identical names
func (t TypeTuple) Equals(other Type) bool {
	if t2, ok := other.(TypeTuple); ok {
		if len(t.Children) != len(t2.Children) {
			return false
		}

		for i, child := range t.Children {
			child2 := t2.Children[i]
			if child.Equals(child2) == false {
				return false
			}
		}

		return true
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t TypeTuple) IsError() bool  { return false }
func (t TypeTuple) String() string { return fmt.Sprintf("(%s)", concatTypes(t.Children)) }
func (t TypeTuple) isType()        {}

// TypeList describes an array of a common type
type TypeList struct {
	Child Type
}

// Equals returns true if another type has an identical structure and identical names
func (t TypeList) Equals(other Type) bool {
	if t2, ok := other.(TypeList); ok {
		return t.Child.Equals(t2.Child)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t TypeList) IsError() bool  { return false }
func (t TypeList) String() string { return fmt.Sprintf("[%s]", t.Child) }
func (t TypeList) isType()        {}

// TypeOptional describes a type that may resolve to a value or nothing
type TypeOptional struct {
	Child Type
}

// Equals returns true if another type has an identical structure and identical names
func (t TypeOptional) Equals(other Type) bool {
	if t2, ok := other.(TypeOptional); ok {
		return t.Child.Equals(t2.Child)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t TypeOptional) IsError() bool  { return false }
func (t TypeOptional) String() string { return fmt.Sprintf("%s?", t.Child) }
func (t TypeOptional) isType()        {}

// TypeIdent describes a type aliased to an identifier
type TypeIdent struct {
	Name string
}

// Equals returns true if another type has an identical structure and identical names
func (t TypeIdent) Equals(other Type) bool {
	if t2, ok := other.(TypeIdent); ok {
		return t.Name == t2.Name
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t TypeIdent) IsError() bool  { return false }
func (t TypeIdent) String() string { return t.Name }
func (t TypeIdent) isType()        {}

func concatTypes(types []Type) string {
	out := ""
	for i, typ := range types {
		if i > 0 {
			out += " "
		}
		out += typ.String()
	}
	return out
}
