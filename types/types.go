package types

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

// Equals returns true for every other type
func (te TypeError) Equals(other Type) bool {
	return false
}

// IsError returns true because TypeError is an error
func (te TypeError) IsError() bool  { return true }
func (te TypeError) String() string { return "ERROR" }
func (te TypeError) isType()        {}

// TypeVoid descirbes the return type of a function that returns no value
type TypeVoid struct{}

// Equals returns true if another type has an identical structure and identical names
func (tv TypeVoid) Equals(other Type) bool {
	if _, ok := other.(TypeVoid); ok {
		return true
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (tv TypeVoid) IsError() bool  { return false }
func (tv TypeVoid) String() string { return "Void" }
func (tv TypeVoid) isType()        {}

// TypeFunction describes mappings of 0+ parameter types to a return type
type TypeFunction struct {
	Params TypeTuple
	Ret    Type
}

// Equals returns true if another type has an identical structure and identical names
func (tf TypeFunction) Equals(other Type) bool {
	if tf2, ok := other.(TypeFunction); ok {
		return tf.Params.Equals(tf2.Params) && tf.Ret.Equals(tf2.Ret)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (tf TypeFunction) IsError() bool  { return false }
func (tf TypeFunction) String() string { return fmt.Sprintf("%s => %s", tf.Params, tf.Ret) }
func (tf TypeFunction) isType()        {}

// TypeTuple describes a group of types
type TypeTuple struct {
	Children []Type
}

// Equals returns true if another type has an identical structure and identical names
func (tt TypeTuple) Equals(other Type) bool {
	if tt2, ok := other.(TypeTuple); ok {
		if len(tt.Children) != len(tt2.Children) {
			return false
		}

		for i, child := range tt.Children {
			child2 := tt2.Children[i]
			if child.Equals(child2) == false {
				return false
			}
		}

		return true
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (tt TypeTuple) IsError() bool  { return false }
func (tt TypeTuple) String() string { return fmt.Sprintf("(%s)", concatTypes(tt.Children)) }
func (tt TypeTuple) isType()        {}

// TypeList describes an array of a common type
type TypeList struct {
	Child Type
}

// Equals returns true if another type has an identical structure and identical names
func (tl TypeList) Equals(other Type) bool {
	if tl2, ok := other.(TypeList); ok {
		return tl.Child.Equals(tl2.Child)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (tl TypeList) IsError() bool  { return false }
func (tl TypeList) String() string { return fmt.Sprintf("[%s]", tl.Child) }
func (tl TypeList) isType()        {}

// TypeOptional describes a type that may resolve to a value or nothing
type TypeOptional struct {
	Child Type
}

// Equals returns true if another type has an identical structure and identical names
func (to TypeOptional) Equals(other Type) bool {
	if to2, ok := other.(TypeOptional); ok {
		return to.Child.Equals(to2.Child)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (to TypeOptional) IsError() bool  { return false }
func (to TypeOptional) String() string { return fmt.Sprintf("%s?", to.Child) }
func (to TypeOptional) isType()        {}

// TypeIdent describes a type aliased to an identifier
type TypeIdent struct {
	Name string
}

// Equals returns true if another type has an identical structure and identical names
func (ti TypeIdent) Equals(other Type) bool {
	if ti2, ok := other.(TypeIdent); ok {
		return ti.Name == ti2.Name
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (ti TypeIdent) IsError() bool  { return false }
func (ti TypeIdent) String() string { return ti.Name }
func (ti TypeIdent) isType()        {}

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
