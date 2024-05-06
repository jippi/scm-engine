//go:build tools
// +build tools

package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/davecgh/go-spew/spew"
	_ "github.com/fatih/structtag"         // needed for go:generate
	_ "github.com/iancoleman/strcase"      // needed for go:generate
	_ "github.com/vektah/gqlparser/v2/ast" // needed for go:generate
)
