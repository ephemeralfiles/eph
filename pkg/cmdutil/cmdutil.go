package cmdutil

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// ValidateRequired checks if a required value is provided and exits if not
func ValidateRequired(value, name string, cmd *cobra.Command) {
	if value == "" {
		fmt.Fprintf(os.Stderr, "%s is required\n", name)
		_ = cmd.Usage()
		os.Exit(1)
	}
}

// HandleError prints an error message and exits with code 1
func HandleError(message string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", message, err)
	os.Exit(1)
}

// HandleErrorf prints a formatted error message and exits with code 1
func HandleErrorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}