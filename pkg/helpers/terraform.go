package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	tJson "github.com/hashicorp/hcl/v2/json"
)

type TerraformConfig struct {
	Terraform TerraformBlock `hcl:"terraform,block"`
}

type TerraformBlock struct {
	Backend TerraformBackendBlock `hcl:"backend,block"`
}

type TerraformBackendBlock struct {
	Type         string      `hcl:"type,label"`
	Organization string      `hcl:"organization,attr"`
	Workspaces   []Workspace `hcl:"workspaces,block"`
}

type Workspace struct {
	Name string `hcl:"name"`
}

func FindBackend(rootDir string) (*TerraformConfig, error) {
	var config *TerraformConfig
	sanitizedRootDir := filepath.Clean(rootDir)
	items, err := ioutil.ReadDir(sanitizedRootDir)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		path := path.Join(sanitizedRootDir, item.Name())
		if item.IsDir() {
			continue
		}

		config, err = processFile(path)
		if err != nil {
			return nil, err
		}
	}
	if config == nil {
		return nil, errors.New("was not able to find a remote backend configuration")
	}
	if config.Terraform.Backend.Type != "remote" {
		return nil, fmt.Errorf("found backend with type %s, expected remote", config.Terraform.Backend.Type)
	}
	return config, nil
}

func FindTerraformToken(hostUrl string) (string, error) {
	u, err := url.Parse(hostUrl)
	if err != nil {
		return "", err
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	tfrcPath := path.Join(dir, ".terraform.d", "credentials.tfrc.json")
	file, err := ioutil.ReadFile(tfrcPath)
	if err != nil {
		return "", err
	}
	var result map[string]interface{}
	err = json.Unmarshal(file, &result)
	if err != nil {
		return "", err
	}
	credentials, exists := result["credentials"]
	if !exists {
		return "", fmt.Errorf("could not find credentials for %s", u.Host)
	}
	host, exists := credentials.(map[string]interface{})[u.Host]
	if !exists {
		return "", fmt.Errorf("could not find credentials for %s", u.Host)
	}
	token, exists := host.(map[string]interface{})["token"]
	if !exists {
		return "", fmt.Errorf("could not find credentials for %s", u.Host)
	}
	ret, ok := token.(string)
	if !ok {
		return "", fmt.Errorf("could not find credentials for %s", u.Host)
	}
	return ret, nil
}

func processFile(path string) (*TerraformConfig, error) {
	file, err := parseFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var config TerraformConfig
	diag := gohcl.DecodeBody(file.Body, nil, &config)
	if diag.HasErrors() {
		errors := []error{}
		for _, err := range diag.Errs() {
			if strings.Contains(err.Error(), "Unsupported block type") {
				continue
			}
			if strings.Contains(err.Error(), "Unsupported argument") {
				continue
			}
			if strings.Contains(err.Error(), "Extraneous JSON object property") {
				continue
			}
			errors = append(errors, err)
		}
		if len(errors) > 0 {
			return nil, fmt.Errorf("%+v", errors)
		}

	}

	return &config, nil
}

func parseFile(path string) (*hcl.File, error) {
	if strings.HasSuffix(path, ".json") {
		return parseJsonFile(path)
	}
	if strings.HasSuffix(path, ".tf") {
		return parseHCLFile(path)
	}
	return nil, fmt.Errorf("unsupported file type")
}

func parseJsonFile(path string) (*hcl.File, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file, diag := tJson.Parse(content, path)
	if diag.HasErrors() {
		return nil, diag
	}

	return file, nil
}

func parseHCLFile(path string) (*hcl.File, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file, diag := hclsyntax.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
	if diag.HasErrors() {
		return nil, diag
	}

	return file, nil
}
