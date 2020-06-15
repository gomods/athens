package cmd

import (
	"fmt"
	"os"
)

func errLog(fmtString string, args ...interface{}) {
	fmt.Printf(fmtString, args...)
	os.Exit(1)
}
