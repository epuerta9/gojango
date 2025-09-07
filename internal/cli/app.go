package cli

import (
	"github.com/spf13/cobra"
	"github.com/epuerta9/gojango/internal/cli/commands"
)

// NewApp creates the root CLI application
func NewApp(version, commit, date string) *cobra.Command {
	app := &cobra.Command{
		Use:   "gojango",
		Short: "The Django-like framework for Go",
		Long: `Gojango is a batteries-included web framework for Go that brings
Django's incredible developer experience to Go's performance and type safety.

Create new projects, generate code, and manage your Gojango applications
with this command-line interface.`,
		Version: version,
		Example: `  # Create a new project
  gojango new myblog

  # Create a new app (run from within project)
  gojango startapp blog

  # Show version
  gojango version`,
	}

	// Add global flags
	app.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	app.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-essential output")

	// Add commands
	app.AddCommand(commands.NewProjectCmd())
	app.AddCommand(commands.NewStartAppCmd())
	app.AddCommand(commands.NewGenerateCmd())
	app.AddCommand(commands.NewVersionCmd(version, commit, date))
	app.AddCommand(commands.NewDoctorCmd())

	// Set custom help template
	app.SetHelpTemplate(helpTemplate)

	return app
}

const helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`