package config

import (
	"bytes"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFile loads and parses a GITLAB_LABELS file at the path specified.
func LoadFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return ParseFile(f)
}

// ParseFile parses a Gitlabber file, returning a Config.
func ParseFile(f io.Reader) (*Config, error) {
	config := &Config{}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(f); err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(buf.Bytes(), config); err != nil {
		return nil, err
	}

	return config, nil
}

// ParseFile parses a Gitlabber file, returning a Config.
func ParseFileString(in string) (*Config, error) {
	config := &Config{}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(strings.NewReader((in))); err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(buf.Bytes(), config); err != nil {
		return nil, err
	}

	return config, nil
}
