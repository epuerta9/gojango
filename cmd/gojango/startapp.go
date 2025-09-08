package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newStartAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "startapp [app-name]",
		Short: "Create a new Django-style app",
		Long: `Create a new application in the apps/ directory with Django-style structure.

This creates:
- app.go (app configuration and routes)
- schema/ (Ent schemas)
- templates/ (HTML templates)
- static/ (static files)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appName := args[0]
			appPath := filepath.Join("apps", appName)

			fmt.Printf("Creating app '%s' in %s...\n", appName, appPath)

			// Create app directory structure
			dirs := []string{
				appPath,
				filepath.Join(appPath, "schema"),
				filepath.Join(appPath, "templates", appName),
				filepath.Join(appPath, "static", appName),
			}

			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", dir, err)
				}
			}

			// Generate app.go
			appGoContent := fmt.Sprintf(`package %s

import (
	"github.com/epuerta9/gojango/pkg/gojango"
	"github.com/gin-gonic/gin"
)

func init() {
	gojango.Register(&%sApp{})
}

type %sApp struct{}

func (app *%sApp) Config() gojango.AppConfig {
	return gojango.AppConfig{
		Name:  "%s",
		Label: "%s Application",
	}
}

func (app *%sApp) Initialize(ctx *gojango.AppContext) error {
	return nil
}

func (app *%sApp) Routes() []gojango.Route {
	return []gojango.Route{
		{
			Method:  "GET",
			Path:    "/%s/",
			Handler: app.IndexView,
			Name:    "%s:index",
		},
	}
}

func (app *%sApp) IndexView(c *gin.Context) {
	c.HTML(200, "%s/index.html", gin.H{
		"title": "%s",
	})
}
`, appName, 
   capitalize(appName), 
   capitalize(appName), 
   capitalize(appName), 
   appName, 
   capitalize(appName), 
   capitalize(appName), 
   capitalize(appName), 
   appName, 
   appName, 
   capitalize(appName), 
   appName, 
   capitalize(appName))

			if err := os.WriteFile(filepath.Join(appPath, "app.go"), []byte(appGoContent), 0644); err != nil {
				return fmt.Errorf("failed to create app.go: %w", err)
			}

			// Generate template
			templateContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
</head>
<body>
    <h1>%s App</h1>
    <p>Welcome to the %s application!</p>
</body>
</html>
`, capitalize(appName), appName)

			templatePath := filepath.Join(appPath, "templates", appName, "index.html")
			if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
				return fmt.Errorf("failed to create template: %w", err)
			}

			fmt.Printf(`✅ Successfully created app '%s'

App structure:
  %s/
  ├── app.go          # App configuration and views
  ├── schema/         # Ent schemas
  ├── templates/%s/   # HTML templates
  └── static/%s/      # Static files

Don't forget to add "%s" to your INSTALLED_APPS in config/settings.star
`, appName, appPath, appName, appName, "apps."+appName)

			return nil
		},
	}

	return cmd
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}