package validators

import (
	"fmt"
	"time"

	"github.com/gobuffalo/validate"
)

type TimeIsPresent struct {
	Name    string
	Field   time.Time
	Message string
}

// IsValid adds an error if the field is not a valid time.
func (v *TimeIsPresent) IsValid(errors *validate.Errors) {
	t := time.Time{}
	if v.Field.UnixNano() != t.UnixNano() {
		return
	}

	if len(v.Message) > 0 {
		errors.Add(GenerateKey(v.Name), v.Message)
		return
	}

	errors.Add(GenerateKey(v.Name), fmt.Sprintf("%s can not be blank.", v.Name))
}
