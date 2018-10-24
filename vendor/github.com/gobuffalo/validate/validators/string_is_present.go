package validators

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/validate"
)

type StringIsPresent struct {
	Name    string
	Field   string
	Message string
}

// IsValid adds an error if the field is empty.
func (v *StringIsPresent) IsValid(errors *validate.Errors) {
	if strings.TrimSpace(v.Field) != "" {
		return
	}

	if len(v.Message) > 0 {
		errors.Add(GenerateKey(v.Name), v.Message)
		return
	}

	errors.Add(GenerateKey(v.Name), fmt.Sprintf("%s can not be blank.", v.Name))
}
