package action

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/hashicorp/go-tfe"
)

type Upload struct {
	Organization string
	Workspace    string
	Token        string
	Url          string
}

func NewUpload() *Upload {
	return &Upload{}
}

func (u *Upload) Run(path string) error {
	config := &tfe.Config{
		Address: u.Url,
		Token:   u.Token,
	}
	ctx := context.Background()

	client, err := tfe.NewClient(config)
	if err != nil {
		return err
	}
	workspace, err := client.Workspaces.Read(ctx, u.Organization, u.Workspace)
	if err != nil {
		return err
	}

	configVersion, err := client.ConfigurationVersions.Create(ctx, workspace.ID, tfe.ConfigurationVersionCreateOptions{
		AutoQueueRuns: tfe.Bool(false),
	})
	if err != nil {
		return err
	}
	if err := client.ConfigurationVersions.Upload(ctx, configVersion.UploadURL, path); err != nil {
		return err
	}

	message, err := getHeadCommitMessage(path)
	if err != nil {
		return err
	}

	client.Runs.Create(ctx, tfe.RunCreateOptions{
		Type:                 "run",
		Workspace:            workspace,
		ConfigurationVersion: configVersion,
		Message:              tfe.String(message),
	})

	return nil
}

func getHeadCommitMessage(path string) (string, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", err
	}
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", err
	}
	return commit.Message, err
}
