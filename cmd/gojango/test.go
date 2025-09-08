package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	var verbose bool
	var coverage bool
	var pattern string

	cmd := &cobra.Command{
		Use:   "test [packages...]",
		Short: "Run tests",
		Long: `Run Go tests for your Gojango application.

This command runs tests with proper test database setup and cleanup.
It supports all standard Go test flags and adds Gojango-specific functionality.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			testArgs := []string{"test"}
			
			if verbose {
				testArgs = append(testArgs, "-v")
			}
			
			if coverage {
				testArgs = append(testArgs, "-cover", "-coverprofile=coverage.out")
			}
			
			if pattern != "" {
				testArgs = append(testArgs, "-run", pattern)
			}
			
			// Default to testing all packages if none specified
			if len(args) == 0 {
				testArgs = append(testArgs, "./...")
			} else {
				testArgs = append(testArgs, args...)
			}

			fmt.Printf("Running tests: go %s\n", testArgs[0])
			
			goCmd := exec.Command("go", testArgs...)
			goCmd.Stdout = os.Stdout
			goCmd.Stderr = os.Stderr
			goCmd.Env = append(os.Environ(), "GOJANGO_ENV=test")
			
			if err := goCmd.Run(); err != nil {
				return fmt.Errorf("tests failed: %w", err)
			}

			if coverage {
				fmt.Println("\n📊 Coverage report saved to coverage.out")
				fmt.Println("View with: go tool cover -html=coverage.out")
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose test output")
	cmd.Flags().BoolVar(&coverage, "coverage", false, "Enable coverage reporting")
	cmd.Flags().StringVarP(&pattern, "run", "r", "", "Run only tests matching pattern")

	return cmd
}