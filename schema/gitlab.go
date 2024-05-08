//go:build ignore

package main

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin/modelgen"
	"github.com/fatih/structtag"
	"github.com/iancoleman/strcase"
	"github.com/vektah/gqlparser/v2/ast"
)

//go:embed docs.tmpl
var docs string

var (
	Props   = []*Property{}
	PropMap = map[string]*Property{}
)

func main() {
	PropMap = make(map[string]*Property)

	cfg, err := config.LoadConfig(getRootPath() + "/schema/gitlab.gqlgen.yml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config", err.Error())

		os.Exit(2)
	}

	// Attaching the mutation function onto model-gen plugin
	p := modelgen.Plugin{
		FieldHook:  fieldHook,
		MutateHook: mutateHook,
	}

	err = api.Generate(cfg, api.ReplacePlugin(&p))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}

	nest(Props)

	var index bytes.Buffer
	tmpl := template.Must(template.New("index").Parse(docs))

	if err := tmpl.Execute(&index, Props[0]); err != nil {
		panic(err)
	}

	fmt.Println(index.String())
}

func nest(props []*Property) {
	for _, field := range props {
		if field.IsCustomType {
			attr, ok := PropMap[field.Type]
			if !ok {
				continue
			}

			for _, nested := range attr.Attributes {
				field.AddAttribute(&Property{
					Name:         nested.Name,
					Description:  nested.Description,
					Optional:     nested.Optional,
					Type:         nested.Type,
					IsSlice:      nested.IsSlice,
					IsCustomType: nested.IsCustomType,
				})
			}
		}

		nest(field.Attributes)
	}
}

func getRootPath() string {
	path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(path))
}

// Defining mutation function
func fieldHook(td *ast.Definition, fd *ast.FieldDefinition, f *modelgen.Field) (*modelgen.Field, error) {
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
		modelName := model.Name

		if modelName != "Context" {
			modelName = strings.TrimPrefix(modelName, "Context")
		}

		modelName = strcase.ToSnake(modelName)

		modelProperty := &Property{
			Name:        modelName,
			Type:        modelName,
			Description: model.Description,
		}

		fmt.Println("model", modelProperty.Name)

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

			exprTags, err := tags.Get("expr")
			if err != nil {
				panic(err)
			}

			if exprTags.Name != "-" {
				fieldType := field.Type.String()

				fieldProperty := &Property{
					Name:        exprTags.Name,
					Optional:    field.Omittable || strings.HasPrefix(fieldType, "*"),
					IsSlice:     strings.HasPrefix(fieldType, "[]"),
					Description: field.Description,
				}

				if strings.Contains(fieldType, "github.com/jippi/scm-engine") {
					fieldType = filepath.Base(fieldType)
					fieldType = strings.Split(fieldType, ".")[1]
					fieldType = strings.TrimPrefix(fieldType, "Context")
					fieldType = strcase.ToSnake(fieldType)

					fieldProperty.IsCustomType = true
				}

				switch {
				case strings.Contains(fieldType, "time.Time"):
					fieldType = "time"

				case strings.Contains(fieldType, "time.Duration"):
					fieldType = "duration"
				}

				fieldProperty.Type = strings.TrimPrefix(fieldType, "*")

				modelProperty.AddAttribute(fieldProperty)
			} // end expr tag is set

			slices.SortFunc(modelProperty.Attributes, sortSlice)

			field.Tag = tags.String()
		} // end fields loop

		if strings.HasSuffix(model.Name, "Node") || model.Name == "Query" {
			continue
		}

		Props = append(Props, modelProperty)
		PropMap[modelProperty.Type] = modelProperty
	} // end model loop

	return b
}

// Property represents either a HCL block (with its sub-blocks or sub-attributes)
// or a single attribute (with no child nodes)
type Property struct {
	// Name of the property (e.g. "merge_request")
	Name        string
	Description string

	// Is the property optional?
	Optional bool

	// The underlying type of the field (e.g. "string", "int", etc.)
	Type string

	// Tracks if this property is a slice (wether its a list of blocks or a list of a scalar type).
	// Used to show "String list" or "Block list" in the documentation output
	IsSlice bool

	IsCustomType bool

	// Contains any sub-attributes for this Property.
	Attributes []*Property

	// Used to track the hierarchy of properties - for example to compute the filename for external
	// markdown documentation for the [Usage] field.
	Parent *Property
}

func (p *Property) AddAttribute(attrs ...*Property) {
	for _, attr := range attrs {
		if attr == nil {
			return
		}

		attr.Parent = p

		p.Attributes = append(p.Attributes, attr)
	}
}

// getHierarchy returns a slice representing all ancestors of this Property
// and its own Property name
func (p Property) getHierarchy() []string {
	// This ensure the "root" node called [project] is not included in the hierarchy
	if p.Parent == nil {
		return nil
	}

	out := []string{}

	if p.Parent != nil {
		out = append(out, p.Parent.getHierarchy()...)
	}

	name := p.Name
	if p.IsSlice && p.IsCustomType {
		name += "[]"
	}

	return append(out, name)
}

func (p *Property) BlockName() string {
	if h := p.getHierarchy(); len(h) > 1 {
		return strings.Join(h, ".")
	}

	return p.Name
}

func sortSlice(i, j *Property) int {
	return cmp.Compare(i.Name, j.Name)
}
