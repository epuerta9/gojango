package version

import "fmt"

// Build-time variables (set by ldflags)
var (
	// Version is the current version of Gojango
	Version = "0.2.0"
	// Commit is the git commit hash
	Commit = "unknown"
	// Date is the build date
	Date = "unknown"
)

// Info holds version information
type Info struct {
	Version string
	Commit  string
	Date    string
}

// Get returns the current version information
func Get() Info {
	return Info{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	if i.Commit != "unknown" && i.Date != "unknown" {
		return fmt.Sprintf("%s (commit: %s, built: %s)", i.Version, i.Commit[:8], i.Date)
	}
	return i.Version
}

// CLIString returns a CLI-formatted version string
func (i Info) CLIString() string {
	return fmt.Sprintf("Gojango CLI version %s\nFramework: github.com/epuerta9/gojango\nDocumentation: https://github.com/epuerta9/gojango\nRepository: https://github.com/epuerta9/gojango", i.String())
}

// AppString returns an application-formatted version string
func (i Info) AppString() string {
	return fmt.Sprintf("Gojango application version %s", i.String())
}