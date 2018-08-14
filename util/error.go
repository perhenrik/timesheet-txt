package util

import (
	"fmt"
	"os"
)

// Check if we have an error and prints a message to stderr
func Check(e error) {
	if e != nil {
		_, err := fmt.Fprintf(os.Stderr, "error: %s\n", e.Error())
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}
