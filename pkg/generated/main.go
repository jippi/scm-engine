package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/jippi/scm-engine/pkg/config"
)

func main() {
	r := new(jsonschema.Reflector)
	if err := r.AddGoComments("github.com/jippi/scm-engine", "./pkg"); err != nil {
		panic(err)
	}

	schema := r.Reflect(&config.Config{})

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(dir+"/pkg/generated/resources/scm-engine.schema.json", data, 0o600); err != nil {
		panic(err)
	}
}
