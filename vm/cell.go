package vm

import (
	"fmt"
)

// CellTemplate is used in the codegen stage to build relationships between
// uses of the same variable
type CellTemplate struct {
	ID   uint
	Name string
}

func (ct *CellTemplate) String() string { return fmt.Sprintf("%s<%d>", ct.Name, ct.ID) }

// Cell holds the binding between a variable and the object it points to
type Cell struct {
	ID  uint
	Ref Object
}

func (c *Cell) String() string { return fmt.Sprintf("<%d>", c.ID) }
