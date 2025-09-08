package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	var port string
	var debug bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the development server",
		Long: `Start the Gojango development server.

This command starts the server with hot reloading and debug mode enabled.
It will look for the main application entry point and start the server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Look for server entry point
			serverPath := "cmd/server/main.go"
			if _, err := os.Stat(serverPath); os.IsNotExist(err) {
				serverPath = "main.go"
			}

			if _, err := os.Stat(serverPath); os.IsNotExist(err) {
				return fmt.Errorf("no server entry point found (tried cmd/server/main.go and main.go)")
			}

			// Set environment variables
			if debug {
				os.Setenv("DEBUG", "true")
			}
			if port != "" {
				os.Setenv("PORT", port)
			}

			fmt.Printf("Starting server from %s...\n", serverPath)
			
			// Run with go run for development
			goCmd := exec.Command("go", "run", serverPath)
			goCmd.Stdout = os.Stdout
			goCmd.Stderr = os.Stderr
			goCmd.Stdin = os.Stdin
			
			return goCmd.Run()
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run server on")
	cmd.Flags().BoolVar(&debug, "debug", true, "Enable debug mode")

	return cmd
}