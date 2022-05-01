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
			fmt.Println("Version:", version.Version)
			fmt.Println("Commit:", version.CommitSha)
			fmt.Println("Build date:", version.BuildDate)
			return nil
		},
	}
}
