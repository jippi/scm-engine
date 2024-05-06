//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin/modelgen"
	"github.com/fatih/structtag"
	"github.com/iancoleman/strcase"
	"github.com/vektah/gqlparser/v2/ast"
)

// Defining mutation function
func constraintFieldHook(td *ast.Definition, fd *ast.FieldDefinition, f *modelgen.Field) (*modelgen.Field, error) {
	// Call default hook to proceed standard directives like goField and goTag.
	// You can omit it, if you don't need.
	if f, err := modelgen.DefaultFieldMutateHook(td, fd, f); err != nil {
		return f, err
	}

	tags, err := structtag.Parse(f.Tag)
	if err != nil {
		return nil, err
	}

	// Remove JSON tag, we don't need it
	tags.Delete("json")

	if c := fd.Directives.ForName("internal"); c != nil {
		tags.Set(&structtag.Tag{Key: "expr", Name: "-"})
	} else if c := fd.Directives.ForName("expr"); c != nil {
		value := c.Arguments.ForName("key")

		if value != nil {
			tags.Set(&structtag.Tag{Key: "expr", Name: value.Value.Raw})
		}
	}

	if c := fd.Directives.ForName("generated"); c != nil {
		tags.Set(&structtag.Tag{Key: "graphql", Name: "-"})
	} else if c := fd.Directives.ForName("graphql"); c != nil {
		value := c.Arguments.ForName("key")

		if value != nil {
			tags.Set(&structtag.Tag{Key: "graphql", Name: value.Value.Raw})
		}
	}

	f.Tag = tags.String()

	return f, nil
}

func mutateHook(b *modelgen.ModelBuild) *modelgen.ModelBuild {
	for _, model := range b.Models {
		for _, field := range model.Fields {
			tags, err := structtag.Parse(field.Tag)
			if err != nil {
				return b
			}

			if !strings.Contains(field.Tag, "expr:") {
				tags.Set(&structtag.Tag{Key: "expr", Name: strcase.ToSnake(field.Name)})
			}

			if !strings.Contains(field.Tag, "graphql:") {
				tags.Set(&structtag.Tag{Key: "graphql", Name: strcase.ToLowerCamel(field.Name)})
			}

			field.Tag = tags.String()
		}
	}

	return b
}

func main() {
	cfg, err := config.LoadConfig(getRootPath() + "/schema/gitlab.gqlgen.yml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config", err.Error())

		os.Exit(2)
	}

	// Attaching the mutation function onto modelgen plugin
	p := modelgen.Plugin{
		FieldHook:  constraintFieldHook,
		MutateHook: mutateHook,
	}

	err = api.Generate(cfg, api.ReplacePlugin(&p))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}
}

func getRootPath() string {
	path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(path))
}
