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
			upload.Path = args[0]
			return upload.Run()
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&upload.Organization, "organization", "", "", "Organization")
	flags.StringVarP(&upload.Workspace, "workspace", "", "", "Workspace")
	flags.StringVarP(&upload.Token, "token", "t", "", "Token")
	flags.StringVarP(&upload.Url, "url", "u", "", "URL")
	flags.StringVarP(&upload.Message, "message", "m", "", "Message")
	flags.BoolVarP(&upload.Apply, "apply", "a", false, "Apply")
	flags.BoolVarP(&upload.Wait, "wait", "w", false, "Wait")

	return cmd
}
