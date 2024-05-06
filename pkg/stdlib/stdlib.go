package stdlib

import "github.com/expr-lang/expr"

var Functions = []expr.Option{
	// Replace built-in duration function with one that supports "d" (days) and "w" (weeks)
	expr.DisableBuiltin("duration"),
	Duration,

	// filepath.Dir
	FilepathDir,

	// some/deep/path/ok => some/deep
	LimitPathDepthTo,

	// slices.Sort + slices.Compact
	Uniq,
}
