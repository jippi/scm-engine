package cmd

import (
	"encoding/json"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/kaptinlin/jsonschema"
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

	raw, err := os.ReadFile(state.ConfigFilePath(ctx))
	if err != nil {
		return err
	}

	var out map[string]any
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return err
	}

	schema, err := jsonschema.NewCompiler().GetSchema("https://jippi.github.io/scm-engine/scm-engine.schema.json")
	if err != nil {
		return err
	}

	result := schema.Validate(out)

	details, _ := json.MarshalIndent(result.ToList(true), "", "  ")
	log.Println(string(details))

	dummyContext := gitlab.Context{}
	spew.Dump(cfg.Lint(ctx, &dummyContext))

	return nil
}
