// Gojango CLI - The global command-line interface for Gojango framework
//
// This CLI tool provides project scaffolding and code generation capabilities.
// It is installed globally and used to create new projects and apps.
package main

import (
	"fmt"
	"os"

	"github.com/epuerta9/gojango/internal/cli"
)

// Build-time variables (set by GoReleaser)
var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	app := cli.NewApp(version)
	
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}