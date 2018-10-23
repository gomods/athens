package validators

import (
	"fmt"

	"github.com/gobuffalo/validate"
)

type IntIsLessThan struct {
	Name     string
	Field    int
	Compared int
	Message  string
}

// IsValid adds an error if the field is not less than the compared value.
func (v *IntIsLessThan) IsValid(errors *validate.Errors) {
	if v.Field < v.Compared {
		return
	}

	if len(v.Message) > 0 {
		errors.Add(GenerateKey(v.Name), v.Message)
		return
	}

	errors.Add(GenerateKey(v.Name), fmt.Sprintf("%d is not less than %d.", v.Field, v.Compared))
}
