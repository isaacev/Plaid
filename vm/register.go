package vm

var regIDCounter int

// RegisterTemplate is bound to each unique variable during the codegen stage
// where each template has a unique id
type RegisterTemplate struct {
	ID   int
	Name string
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
