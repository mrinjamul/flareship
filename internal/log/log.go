package log

import (
	"fmt"
	"os"
)

var verbose bool

// SetVerbose sets the verbosity level for logging.
func SetVerbose(v bool) {
	verbose = v
}

// Info prints informational messages.
func Info(format string, a ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", a...)
}

// Error prints error messages and exits.
func Error(format string, a ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", a...)
	os.Exit(1)
}

// Debug prints debug messages if verbose mode is enabled.
func Debug(format string, a ...interface{}) {
	if verbose {
		fmt.Printf("[DEBUG] "+format+"\n", a...)
	}
}
