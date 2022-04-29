package action

import (
	"context"
	"fmt"
	"os"
	"time"

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
	Apply        bool
	Wait         bool
}

func NewUpload() *Upload {
	return &Upload{}
}

func (u *Upload) Run() error {
	config := &tfe.Config{
		Address: u.Url,
		Token:   u.Token,
	}
	err := setTokenAndAddress(config)
	if err != nil {
		return err
	}

	message, err := u.getMessage()
	if err != nil {
		return err
	}

	workspaceInfo, err := u.getWorkspaceInformation()
	if err != nil {
		return err
	}
	fmt.Printf("Uploading to %s/%s\n", workspaceInfo.Organization, workspaceInfo.Workspace)

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
		AutoQueueRuns: tfe.Bool(u.Apply),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Created configuration version %s\n", configVersion.ID)

	if err := u.upload(ctx, client, configVersion); err != nil {
		return err
	}

	run, err := client.Runs.Create(ctx, tfe.RunCreateOptions{
		Type:                 "run",
		Workspace:            workspace,
		ConfigurationVersion: configVersion,
		Message:              tfe.String(message),
	})
	if err != nil {
		return err
	}
	fmt.Printf("Created run %s\n", run.ID)
	runUrl := fmt.Sprintf("%s/app/%s/workspaces/%s/runs/%s", config.Address, workspaceInfo.Organization, workspaceInfo.Workspace, run.ID)
	fmt.Printf("Check out the run at: %s\n", runUrl)

	if u.Wait {
		fmt.Println("Waiting for run to apply...")
		if err := waitForRunToApply(ctx, client, run.ID); err != nil {
			return err
		}
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

func setTokenAndAddress(c *tfe.Config) error {
	if c.Token != "" {
		return nil
	}
	defaults := tfe.DefaultConfig()
	if defaults.Token != "" {
		c.Token = defaults.Token
		return nil
	}

	url := c.Address
	if url == "" {
		url = defaults.Address
	}

	token, err := helpers.FindTerraformToken(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: unable to find Terraform token")
		return err
	}
	c.Token = token
	if c.Address == "" {
		c.Address = defaults.Address
	}
	return nil
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

func (u *Upload) upload(ctx context.Context, client *tfe.Client, configVersion *tfe.ConfigurationVersion) error {
	if err := client.ConfigurationVersions.Upload(ctx, configVersion.UploadURL, u.Path); err != nil {
		return err
	}

	fmt.Println("Waiting for upload to complete...")

	for {
		current, err := client.ConfigurationVersions.Read(ctx, configVersion.ID)
		if err != nil {
			return err
		}
		if current.Status == tfe.ConfigurationUploaded {
			break
		}
	}

	return nil
}

func waitForRunToApply(ctx context.Context, client *tfe.Client, runID string) error {
	for {
		run, err := client.Runs.Read(ctx, runID)
		if err != nil {
			return err
		}
		if run.Status == tfe.RunApplied || run.Status == tfe.RunPlannedAndFinished {
			fmt.Printf("Run applied with status: %s\n", run.Status)
			break
		}
		time.Sleep(time.Second)
	}
	return nil
}
