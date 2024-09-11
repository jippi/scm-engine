package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/urfave/cli/v2"
	slogctx "github.com/veqryn/slog-context"
	"gopkg.in/yaml.v3"
)

func Lint(cCtx *cli.Context) error {
	ctx := cCtx.Context
	ctx = state.WithConfigFilePath(ctx, cCtx.String(FlagConfigFile))

	cfg, err := config.LoadFile(state.ConfigFilePath(ctx))
	if err != nil {
		return err
	}

	if len(cfg.Includes) != 0 {
		slogctx.Warn(ctx, "Configuration file contains 'include' settings, those are currently unsupported by 'lint' command and will be ignored")
	}

	// Read raw YAML file
	raw, err := os.ReadFile(state.ConfigFilePath(ctx))
	if err != nil {
		return err
	}

	// Parse the YAML file into lose Go shape
	var yamlOutput any
	if err := yaml.Unmarshal(raw, &yamlOutput); err != nil {
		return err
	}

	// Setup file loaders for reading the JSON schema file
	loader := jsonschema.SchemeURLLoader{
		"file":  jsonschema.FileLoader{},
		"http":  newHTTPURLLoader(),
		"https": newHTTPURLLoader(),
	}

	// Create json schema compiler
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(loader)

	// Compile the schema into validator format
	sch, err := compiler.Compile(cCtx.String("schema"))
	if err != nil {
		return err
	}

	// Validate the json output
	if err := sch.Validate(yamlOutput); err != nil {
		return err
	}

	// To scm-engine specific linting last
	return cfg.Lint(ctx, &gitlab.Context{})
}

type HTTPURLLoader http.Client

func (l *HTTPURLLoader) Load(url string) (any, error) {
	client := (*http.Client)(l)

	resp, err := client.Get(url) //nolint
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()

		return nil, fmt.Errorf("%s returned status code %d", url, resp.StatusCode)
	}

	defer resp.Body.Close()

	return jsonschema.UnmarshalJSON(resp.Body)
}

func newHTTPURLLoader() *HTTPURLLoader {
	httpLoader := HTTPURLLoader(http.Client{
		Timeout: 15 * time.Second,
	})

	return &httpLoader
}
