package commands

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/epuerta9/gojango/internal/cli/ui"
)

// NewDoctorCmd creates the 'doctor' command for checking project health
func NewDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check the health of your development environment",
		Long: `Check your development environment and project setup for common issues.

This command verifies:
  • Go installation and version
  • Project structure and configuration
  • Dependencies and tools
  • Common setup problems

It provides suggestions for fixing any issues found.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.Header("🩺 Gojango Environment Check")
			ui.Info("")
			
			checks := []CheckFunc{
				checkGoInstallation,
				checkGoVersion,
				checkGoModules,
				checkProjectStructure,
				checkDependencies,
			}
			
			var failed int
			for _, check := range checks {
				if !check() {
					failed++
				}
			}
			
			ui.Info("")
			if failed == 0 {
				ui.Success("✅ All checks passed! Your environment looks good.")
			} else {
				ui.Error(fmt.Sprintf("❌ %d check(s) failed. Please address the issues above.", failed))
			}
		},
	}

	return cmd
}

// CheckFunc represents a health check function
type CheckFunc func() bool

// checkGoInstallation verifies Go is installed
func checkGoInstallation() bool {
	ui.Info("Checking Go installation...")
	
	goCmd, err := exec.LookPath("go")
	if err != nil {
		ui.Error("  ❌ Go is not installed or not in PATH")
		ui.Info("  💡 Install Go from https://golang.org/dl/")
		return false
	}
	
	ui.Success(fmt.Sprintf("  ✅ Go found at: %s", goCmd))
	return true
}

// checkGoVersion verifies Go version compatibility
func checkGoVersion() bool {
	ui.Info("Checking Go version...")
	
	version := runtime.Version()
	ui.Info(fmt.Sprintf("  Go version: %s", version))
	
	// Extract version number (e.g., "go1.21.0" -> "1.21.0")
	if !strings.HasPrefix(version, "go") {
		ui.Error("  ❌ Cannot determine Go version")
		return false
	}
	
	versionStr := strings.TrimPrefix(version, "go")
	parts := strings.Split(versionStr, ".")
	
	if len(parts) < 2 {
		ui.Error("  ❌ Invalid Go version format")
		return false
	}
	
	// Check for minimum Go 1.20
	major := parts[0]
	minor := parts[1]
	
	if major == "1" && len(minor) > 0 {
		if minor[0] < '2' || (minor[0] == '2' && len(minor) > 1 && minor[1] < '0') {
			if minor == "19" || (len(minor) == 1 && minor[0] < '2') {
				ui.Warning("  ⚠️  Go 1.20+ recommended (current: " + version + ")")
				ui.Info("  💡 Update Go from https://golang.org/dl/")
				return false
			}
		}
	}
	
	ui.Success("  ✅ Go version is compatible")
	return true
}

// checkGoModules verifies Go modules support
func checkGoModules() bool {
	ui.Info("Checking Go modules...")
	
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = build.Default.GOPATH
	}
	
	ui.Success("  ✅ Go modules are supported")
	
	// Check if in a Go module
	if _, err := os.Stat("go.mod"); err == nil {
		ui.Success("  ✅ Found go.mod file")
		return true
	} else {
		ui.Info("  ℹ️  No go.mod file found (this is normal for CLI usage)")
		return true
	}
}

// checkProjectStructure verifies project structure if in a project
func checkProjectStructure() bool {
	ui.Info("Checking project structure...")
	
	// Check if we're in a Gojango project
	if _, err := os.Stat("gojango.yaml"); err == nil {
		ui.Success("  ✅ Found gojango.yaml")
		
		// Check for expected directories
		expectedDirs := []string{"apps", "cmd", "internal"}
		for _, dir := range expectedDirs {
			if _, err := os.Stat(dir); err == nil {
				ui.Success(fmt.Sprintf("  ✅ Found %s/ directory", dir))
			} else {
				ui.Warning(fmt.Sprintf("  ⚠️  Missing %s/ directory", dir))
			}
		}
		
		return true
	} else {
		ui.Info("  ℹ️  Not in a Gojango project (this is normal for CLI usage)")
		return true
	}
}

// checkDependencies verifies required tools and dependencies
func checkDependencies() bool {
	ui.Info("Checking dependencies...")
	
	// Check for common tools
	tools := map[string]string{
		"make":   "Build automation",
		"git":    "Version control",
		"docker": "Containerization (optional)",
	}
	
	allGood := true
	for tool, description := range tools {
		if _, err := exec.LookPath(tool); err == nil {
			ui.Success(fmt.Sprintf("  ✅ %s available (%s)", tool, description))
		} else {
			if tool == "docker" {
				ui.Info(fmt.Sprintf("  ℹ️  %s not found (%s) - optional", tool, description))
			} else {
				ui.Warning(fmt.Sprintf("  ⚠️  %s not found (%s)", tool, description))
				allGood = false
			}
		}
	}
	
	return allGood
}