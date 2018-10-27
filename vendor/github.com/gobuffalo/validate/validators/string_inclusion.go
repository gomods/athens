package validators

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/validate"
)

type StringInclusion struct {
	Name    string
	Field   string
	List    []string
	Message string
}

// IsValid adds an error if the field is not one of the allowed values.
func (v *StringInclusion) IsValid(errors *validate.Errors) {
	found := false
	for _, l := range v.List {
		if l == v.Field {
			found = true
			break
		}
	}
	if !found {
		if len(v.Message) > 0 {
			errors.Add(GenerateKey(v.Name), v.Message)
			return
		}

		errors.Add(GenerateKey(v.Name), fmt.Sprintf("%s is not in the list [%s].", v.Name, strings.Join(v.List, ", ")))
	}
}
