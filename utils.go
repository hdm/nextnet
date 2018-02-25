package main

import (
	"fmt"
	"os"
	"strings"
)

// PrintVersion prints pretty formatted application version.
func PrintVersion(app string) {
	var version = "master"
	fmt.Fprintf(os.Stderr, "%s v%s\n", app, version)
}

// TrimName deletes all '\x00' symbols from passed name.
func TrimName(name string) string {
	return strings.TrimSpace(strings.Replace(name, "\x00", "", -1))
}
