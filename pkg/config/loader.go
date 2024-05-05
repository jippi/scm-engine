package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFileFromStandardLocation loads and parses a GITLAB_LABELS file at one of the
// standard locations for GITLAB_LABELS files (./, .github/, docs/). If run from a
// git repository, all paths are relative to the repository root.
func LoadFileFromStandardLocation() (*Config, error) {
	path := findFileAtStandardLocation()
	if path == "" {
		return nil, errors.New("could not find GITLAB_LABELS file at any of the standard locations")
	}

	return LoadFile(path)
}

// LoadFile loads and parses a GITLAB_LABELS file at the path specified.
func LoadFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return ParseFile(f)
}

func findFileAtStandardLocation() string {
	pathPrefix := ""

	repoRoot, inRepo := findRepositoryRoot()
	if inRepo {
		pathPrefix = repoRoot
	}

	for _, path := range []string{".scm-engine.yml", ".gitlab/scm-engine.yml", ".github/scm-engine.yml"} {
		fullPath := filepath.Join(pathPrefix, path)

		if fileExists(fullPath) {
			return fullPath
		}
	}

	return ""
}

// fileExist checks if a normal file exists at the path specified.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

// findRepositoryRoot returns the path to the root of the git repository, if
// we're currently in one. If we're not in a git repository, the boolean return
// value is false.
func findRepositoryRoot() (string, bool) {
	output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", false
	}

	return strings.TrimSpace(string(output)), true
}

// ParseFile parses a Gitlabber file, returning a Config.
func ParseFile(f io.Reader) (*Config, error) {
	config := &Config{}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(f); err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(buf.Bytes(), config); err != nil {
		return config, err
	}

	return config, nil
}
