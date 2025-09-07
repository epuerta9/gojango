package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VersionInfo holds build-time version information
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

// NewVersionCmd creates the 'version' command
func NewVersionCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the Gojango version information",
		Long: `Display version information for the Gojango CLI and framework.

This command shows the current version of the Gojango CLI tool and provides
information about the framework version compatibility.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Gojango CLI version %s\n", version)
			fmt.Println("Framework: github.com/epuerta9/gojango")
			fmt.Println("Documentation: https://github.com/epuerta9/gojango")
			fmt.Println("Repository: https://github.com/epuerta9/gojango")
		},
	}

	return cmd
}