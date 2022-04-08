package main

import "github.com/spf13/cobra"

func newTerrapackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terrapack",
		Short: "terrapack",
		Long:  "Terraform Packer and API runner workflow",
	}

	cmd.AddCommand(newVersionCmd())

	return cmd
}
