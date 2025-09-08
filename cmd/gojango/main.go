package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.2.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "gojango",
		Short: "Gojango - Django-inspired web framework for Go",
		Long: `Gojango brings Django's incredible developer experience to Go,
providing batteries-included web development with Go's performance and simplicity.

Features:
- Django-style project structure
- Automatic admin interface from Ent schemas
- Built-in gRPC/Connect APIs with TypeScript generation
- HTMX + Templ for modern server-side rendering
- Multi-app architecture with dependency management
- Code generation over configuration`,
		Version: version,
	}

	// Add global subcommands (Django-admin equivalent)
	rootCmd.AddCommand(newNewCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newCheckCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}