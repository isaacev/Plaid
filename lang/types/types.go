package types

import (
	"fmt"
)

// Type describes all type structs
type Type interface {
	fmt.Stringer
	Equals(other Type) bool
	IsError() bool
	isType()
}

type CompositeType interface {
	Type
	Member(name string) Type
}

type Struct struct {
	Fields []struct {
		Name string
		Type Type
	}
}

func (t Struct) String() (out string) {
	for i, field := range t.Fields {
		out += fmt.Sprintf("%s:%s", field.Name, field.Type)
		if i < len(t.Fields)-1 {
			out += " "
		}
	}

	return "{" + out + "}"
}

func (t Struct) Equals(other Type) bool {
	if t2, ok := other.(Struct); ok {
		if len(t.Fields) != len(t2.Fields) {
			return false
		}

		for i, field := range t.Fields {
			field2 := t2.Fields[i]
			if field.Name != field2.Name || field.Type.Equals(field2.Type) == false {
				return false
			}
		}

		return true
	}

	return false
}

func (t Struct) IsError() bool { return false }
func (t Struct) isType()       {}

func (t Struct) Member(name string) Type {
	for _, field := range t.Fields {
		if field.Name == name {
			return field.Type
		}
	}

	return nil
}

// Error signals that whatever expression was supposed to produce this type
// had a semantic error that made proper evaluation impossible
type Error struct{}

// Equals returns false for every other type
func (t Error) Equals(other Type) bool {
	if _, ok := other.(Error); ok {
		return true
	}

	return false
}

// IsError returns true because TypeError is an error
func (t Error) IsError() bool  { return true }
func (t Error) String() string { return "ERROR" }
func (t Error) isType()        {}

// Any can represent all types
type Any struct{}

// Equals returns true for every other type
func (t Any) Equals(other Type) bool {
	if other.IsError() {
		return false
	} else if _, ok := other.(Void); ok {
		return false
	}

	return true
}

// IsError returns false because this is a properly resolved type
func (t Any) IsError() bool  { return false }
func (t Any) String() string { return "Any" }
func (t Any) isType()        {}

// Void descirbes the return type of a function that returns no value
type Void struct{}

// Equals returns true if another type has an identical structure and identical names
func (t Void) Equals(other Type) bool {
	if _, ok := other.(Void); ok {
		return true
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t Void) IsError() bool  { return false }
func (t Void) String() string { return "Void" }
func (t Void) isType()        {}

// Function describes mappings of 0+ parameter types to a return type
type Function struct {
	Params Tuple
	Ret    Type
}

// Equals returns true if another type has an identical structure and identical names
func (t Function) Equals(other Type) bool {
	if t2, ok := other.(Function); ok {
		return t.Params.Equals(t2.Params) && t.Ret.Equals(t2.Ret)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t Function) IsError() bool  { return false }
func (t Function) String() string { return fmt.Sprintf("%s => %s", t.Params, t.Ret) }
func (t Function) isType()        {}

// Tuple describes a group of types
type Tuple struct {
	Children []Type
}

// Equals returns true if another type has an identical structure and identical names
func (t Tuple) Equals(other Type) bool {
	if t2, ok := other.(Tuple); ok {
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
func (t Tuple) IsError() bool  { return false }
func (t Tuple) String() string { return fmt.Sprintf("(%s)", concatTypes(t.Children)) }
func (t Tuple) isType()        {}

// List describes an array of a common type
type List struct {
	Child Type
}

// Equals returns true if another type has an identical structure and identical names
func (t List) Equals(other Type) bool {
	if t2, ok := other.(List); ok {
		return t.Child.Equals(t2.Child)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t List) IsError() bool  { return false }
func (t List) String() string { return fmt.Sprintf("[%s]", t.Child) }
func (t List) isType()        {}

// Optional describes a type that may resolve to a value or nothing
type Optional struct {
	Child Type
}

// Equals returns true if another type has an identical structure and identical names
func (t Optional) Equals(other Type) bool {
	if t2, ok := other.(Optional); ok {
		return t.Child.Equals(t2.Child)
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t Optional) IsError() bool  { return false }
func (t Optional) String() string { return fmt.Sprintf("%s?", t.Child) }
func (t Optional) isType()        {}

// Ident describes a type aliased to an identifier
type Ident struct {
	Name string
}

// Equals returns true if another type has an identical structure and identical names
func (t Ident) Equals(other Type) bool {
	if t2, ok := other.(Ident); ok {
		return t.Name == t2.Name
	}

	return false
}

// IsError returns false because this is a properly resolved type
func (t Ident) IsError() bool  { return false }
func (t Ident) String() string { return t.Name }
func (t Ident) isType()        {}

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
