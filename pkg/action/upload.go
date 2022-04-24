package action

import (
	"context"

	"github.com/adrianord/terrapack/pkg/helpers"
	"github.com/go-git/go-git/v5"
	"github.com/hashicorp/go-tfe"
)

type Upload struct {
	Organization string
	Workspace    string
	Token        string
	Url          string
	Message      string
	Path         string
}

func NewUpload() *Upload {
	return &Upload{}
}

func (u *Upload) Run() error {
	config := &tfe.Config{
		Address: u.Url,
		Token:   u.Token,
	}

	message, err := u.getMessage()
	if err != nil {
		return err
	}

	workspaceInfo, err := u.getWorkspaceInformation()
	if err != nil {
		return err
	}

	ctx := context.Background()

	client, err := tfe.NewClient(config)
	if err != nil {
		return err
	}

	workspace, err := client.Workspaces.Read(ctx, workspaceInfo.Organization, workspaceInfo.Workspace)
	if err != nil {
		return err
	}

	configVersion, err := client.ConfigurationVersions.Create(ctx, workspace.ID, tfe.ConfigurationVersionCreateOptions{
		AutoQueueRuns: tfe.Bool(false),
	})
	if err != nil {
		return err
	}

	if err := client.ConfigurationVersions.Upload(ctx, configVersion.UploadURL, u.Path); err != nil {
		return err
	}

	_, err = client.Runs.Create(ctx, tfe.RunCreateOptions{
		Type:                 "run",
		Workspace:            workspace,
		ConfigurationVersion: configVersion,
		Message:              tfe.String(message),
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *Upload) getMessage() (string, error) {
	if u.Message != "" {
		return u.Message, nil
	} else {
		return getHeadCommitMessage(u.Path)
	}
}

type workspaceInfo struct {
	Organization string
	Workspace    string
}

func (u *Upload) getWorkspaceInformation() (*workspaceInfo, error) {
	if u.Organization != "" && u.Workspace != "" {
		return &workspaceInfo{
			Organization: u.Organization,
			Workspace:    u.Workspace,
		}, nil
	}

	backend, err := helpers.FindBackend(u.Path)
	if err != nil {
		return nil, err
	}
	if u.Organization != "" {
		return &workspaceInfo{
			Organization: u.Organization,
			Workspace:    backend.Terraform.Backend.Workspaces[0].Name,
		}, nil
	}
	if u.Workspace != "" {
		return &workspaceInfo{
			Organization: backend.Terraform.Backend.Organization,
			Workspace:    u.Workspace,
		}, nil
	}
	return &workspaceInfo{
		Organization: backend.Terraform.Backend.Organization,
		Workspace:    backend.Terraform.Backend.Workspaces[0].Name,
	}, nil
}

func getHeadCommitMessage(path string) (string, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
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
