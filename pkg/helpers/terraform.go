package helpers

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
)

type Config struct {
	Terraform Terraform `hcl:"terraform,block"`
}

type Terraform struct {
	Backend Backend `hcl:"backend,block"`
}

type Backend struct {
	Type         string      `hcl:"type,label"`
	Organization string      `hcl:"organization,attr"`
	Workspaces   []Workspace `hcl:"workspaces,block"`
}

type Workspace struct {
	Name string `hcl:"name"`
}

func FindBackend(rootDir string) (*Config, error) {
	var config *Config
	sanitizedRootDir := filepath.Clean(rootDir)
	err := filepath.WalkDir(sanitizedRootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && strings.Count(path, string(os.PathSeparator)) > 1 {
			return fs.SkipDir
		}

		config, err = processFile(path)
		if err != nil {
			return nil
		}
		if config != nil {
			return io.EOF
		}

		return nil
	})
	if err != nil {
		if err != io.EOF {
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

func processFile(path string) (*Config, error) {
	file, err := parseFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var config Config
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

	file, diag := json.Parse(content, path)
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
