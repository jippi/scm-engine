package stdlib

import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/iancoleman/strcase"
)

var FunctionRenamer = expr.Patch(functionRenamer{})

type functionRenamer struct{}

func (x functionRenamer) Visit(node *ast.Node) {
	switch node := (*node).(type) {
	case *ast.CallNode:
		x.rename(&node.Callee)
	}
}

func (x functionRenamer) rename(node *ast.Node) {
	switch node := (*node).(type) {
	case *ast.MemberNode:
		if !node.Method {
			return
		}

		x.rename(&node.Property)

	case *ast.StringNode:
		node.Value = strcase.ToCamel(node.Value)
	}
}
