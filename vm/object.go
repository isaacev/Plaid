package vm

import "fmt"

// Object represents all objects that can exist in the machine
type Object interface {
	String() string
	isObject()
}

// ObjectInt represents any integer object
type ObjectInt struct {
	val int64
}

func (oi *ObjectInt) String() string { return fmt.Sprintf("%d", oi.val) }
func (oi *ObjectInt) isObject()      {}
