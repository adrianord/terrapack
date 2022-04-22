package main

import (
	"github.com/adrianord/terrapack/pkg/action"
	"github.com/spf13/cobra"
)

func newUploadCmd() *cobra.Command {
	upload := action.NewUpload()
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "upload",
		Long:  "Upload terraform and run API workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return upload.Run(args[0])
		},
	}

	return cmd
}
