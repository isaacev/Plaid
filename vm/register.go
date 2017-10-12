package vm

import "fmt"

var regIDCounter int

// RegisterTemplate is bound to each unique variable during the codegen stage
// where each template has a unique id
type RegisterTemplate struct {
	ID   int
	Name string
}

func (reg *RegisterTemplate) String() string {
	return fmt.Sprintf("%s<%d>", reg.Name, reg.ID)
}

// MakeRegisterTemplate builds a new register template with a unique id and a
// given name
func MakeRegisterTemplate(name string) *RegisterTemplate {
	regIDCounter++
	return &RegisterTemplate{
		ID:   regIDCounter - 1,
		Name: name,
	}
}

// Register holds a value in the virtual machine and is bound to a variable
// during its lifetime. Many registers may be created from a single register
// template
type Register struct {
	Contents Object
	Template *RegisterTemplate
}
