package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show Gojango version",
		Long:  "Display the current version of Gojango.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Gojango %s\n", version)
			return nil
		},
	}

	return cmd
}