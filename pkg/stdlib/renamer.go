package stdlib

import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
)

var FunctionRenamer = expr.Patch(functionRenamer{})

var renames = map[string]string{
	"modified_files": "ModifiedFiles",
}

type functionRenamer struct{}

func (x functionRenamer) Visit(node *ast.Node) {
	switch n := (*node).(type) {
	case *ast.IdentifierNode:
		if r, ok := renames[n.Value]; ok {
			n.Value = r
		}

	case *ast.StringNode:
		if r, ok := renames[n.Value]; ok {
			n.Value = r
		}
	}
}
