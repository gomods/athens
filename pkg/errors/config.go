package errors

import (
	"fmt"
	"strings"
)

// ConfigErrFn is a function that can create a new configuration error. You pass it a
// message specific to the error you found when you were validating configuration,
// and it knows how to print out the actual configuration name and other helpful information.
type ConfigErrFn func(string) error

type configErr struct {
	str string
}

func (c configErr) Error() string {
	return c.str
}

// ConfigError returns a function that creates a configuration error to be printed
// to the terminal. Call this function and pass its return value down the call stack to
// functions that validate configuration fields.
//
// The function that ConfigError returns
//
// For example:
//
//	downloadModeFn := ConfigError("DownloadMode (ATHENS_DOWNLOAD_MODE)")
//	err := doThingsWithDownloadMode(configStruct, downloadModeFn)
func ConfigError(field string, url string) ConfigErrFn {
	return func(helpText string) error {
		slc := []string{
			fmt.Sprintf("There was a configuration error with %s. %s", field, helpText),
		}
		if url != "" {
			slc = append(slc, fmt.Sprintf("See %s for more information.", url))
		}
		return &configErr{str: strings.Join(slc, "\n")}
	}
}
