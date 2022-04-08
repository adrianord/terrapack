package main

import (
	"fmt"

	"github.com/adrianord/terrapack/pkg/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of terrapack",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Terrapack version:", version.Version)
			return nil
		},
	}
}
