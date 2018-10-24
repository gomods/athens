package validators

import (
	"fmt"

	"github.com/gobuffalo/validate"
)

type BytesArePresent struct {
	Name    string
	Field   []byte
	Message string
}

// IsValid adds an error if the field is not empty.
func (v *BytesArePresent) IsValid(errors *validate.Errors) {
	if len(v.Field) > 0 {
		return
	}

	if len(v.Message) > 0 {
		errors.Add(GenerateKey(v.Name), v.Message)
		return
	}

	errors.Add(GenerateKey(v.Name), fmt.Sprintf("%s can not be blank.", v.Name))
}
