// Gojango CLI - The global command-line interface for Gojango framework
//
// This CLI tool provides project scaffolding and code generation capabilities.
// It is installed globally and used to create new projects and apps.
package main

import (
	"fmt"
	"os"

	"github.com/epuerta9/gojango/internal/cli"
	"github.com/epuerta9/gojango/pkg/gojango/version"
)

func init() {
	// Set version information from ldflags at build time
	// Usage: go build -ldflags="-X github.com/epuerta9/gojango/pkg/gojango/version.Version=v0.2.0"
	// This allows the version package to be updated at build time
}

func main() {
	info := version.Get()
	app := cli.NewApp(info.Version, info.Commit, info.Date)
	
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}