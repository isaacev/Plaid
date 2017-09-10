package check

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

// TypeFunction describes mappings of 0+ parameter types to a return type
type TypeFunction struct {
	params TypeTuple
	ret    Type
}

// Equals returns true if another type has an identical structure and identical names
func (tf TypeFunction) Equals(other Type) bool {
	if tf2, ok := other.(TypeFunction); ok {
		return tf.params.Equals(tf2.params) && tf.ret.Equals(tf2.ret)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (tf TypeFunction) IsError() bool  { return false }
func (tf TypeFunction) String() string { return fmt.Sprintf("%s => %s", tf.params, tf.ret) }
func (tf TypeFunction) isType()        {}

// TypeTuple describes a group of types
type TypeTuple struct {
	children []Type
}

// Equals returns true if another type has an identical structure and identical names
func (tt TypeTuple) Equals(other Type) bool {
	if tt2, ok := other.(TypeTuple); ok {
		if len(tt.children) != len(tt2.children) {
			return false
		}

		for i, child := range tt.children {
			child2 := tt2.children[i]
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
func (tt TypeTuple) String() string { return fmt.Sprintf("(%s)", concatTypes(tt.children)) }
func (tt TypeTuple) isType()        {}

// TypeList describes an array of a common type
type TypeList struct {
	child Type
}

// Equals returns true if another type has an identical structure and identical names
func (tl TypeList) Equals(other Type) bool {
	if tl2, ok := other.(TypeList); ok {
		return tl.child.Equals(tl2.child)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (tl TypeList) IsError() bool  { return false }
func (tl TypeList) String() string { return fmt.Sprintf("[%s]", tl.child) }
func (tl TypeList) isType()        {}

// TypeOptional describes a type that may resolve to a value or nothing
type TypeOptional struct {
	child Type
}

// Equals returns true if another type has an identical structure and identical names
func (to TypeOptional) Equals(other Type) bool {
	if to2, ok := other.(TypeOptional); ok {
		return to.child.Equals(to2.child)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (to TypeOptional) IsError() bool  { return false }
func (to TypeOptional) String() string { return fmt.Sprintf("%s?", to.child) }
func (to TypeOptional) isType()        {}

// TypeIdent describes a type aliased to an identifier
type TypeIdent struct {
	name string
}

// Equals returns true if another type has an identical structure and identical names
func (ti TypeIdent) Equals(other Type) bool {
	if ti2, ok := other.(TypeIdent); ok {
		return ti.name == ti2.name
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (ti TypeIdent) IsError() bool  { return false }
func (ti TypeIdent) String() string { return ti.name }
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
