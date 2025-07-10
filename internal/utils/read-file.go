package utils

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func readFileFromGitRoot(filename string) ([]byte, error) {
	gitRoot, err := getGitRoot()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(gitRoot, filename)
	return os.ReadFile(filePath)
}

type JiraBranchConfig struct {
	ProjectKey string `json:"projectKey"`
}

func ReadConfigFile() (JiraBranchConfig, error) {
	file, err := readFileFromGitRoot("jira-branch.config.json")
	if err != nil {
		return JiraBranchConfig{}, err
	}
	var config JiraBranchConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return JiraBranchConfig{}, err
	}
	return config, nil
}
