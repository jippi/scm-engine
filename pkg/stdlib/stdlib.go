package stdlib

import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/patcher/value"
)

var Functions = []expr.Option{
	// Replace built-in duration function with one that supports "d" (days) and "w" (weeks)
	expr.DisableBuiltin("duration"),

	// Add Expr-lang support for a wider range of "valuers" for custom types, such as
	//
	// - "AsString()" interface for custom types wanting to be used as a String (useful for Enum types!)
	// - "AsBool()" interface for custom types wanting to be used as a boolean comparison
	value.ValueGetter,

	Duration,
	Since,

	// filepath.Dir
	FilepathDir,

	// some/deep/path/ok => some/deep
	LimitPathDepthTo,

	// slices.Sort + slices.Compact
	Uniq,
}
