package main

import (
	"os"

	"github.com/spf13/cobra"
)

func newTerrapackCmd() *cobra.Command {
	var cwd string

	cmd := &cobra.Command{
		Use:   "terrapack",
		Short: "terrapack",
		Long:  "Terraform Packer and API runner workflow",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cwd != "" {
				if err := os.Chdir(cwd); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&cwd, "cwd", "C", "", "Run terrapack in a different directory")

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newUploadCmd())

	return cmd
}
