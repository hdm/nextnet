package main

import (
	"fmt"
	"os"
	"strings"
)

func PrintVersion(app string) {
	var version = "master"
	fmt.Fprintf(os.Stderr, "%s v%s\n", app, version)
}

func TrimName(name string) string {
	return strings.TrimSpace(strings.Replace(name, "\x00", "", -1))
}
