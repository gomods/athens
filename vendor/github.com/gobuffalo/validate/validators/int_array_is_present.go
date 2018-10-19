package validators

import (
	"fmt"

	"github.com/gobuffalo/validate"
)

type IntArrayIsPresent struct {
	Name    string
	Field   []int
	Message string
}

// IsValid adds an error if the field is an empty array.
func (v *IntArrayIsPresent) IsValid(errors *validate.Errors) {
	if len(v.Field) > 0 {
		return
	}

	if len(v.Message) > 0 {
		errors.Add(GenerateKey(v.Name), v.Message)
		return
	}

	errors.Add(GenerateKey(v.Name), fmt.Sprintf("%s can not be empty.", v.Name))
}
