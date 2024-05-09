package stdlib

import (
	"cmp"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/xhit/go-str2duration/v2"
)

var FilepathDir = expr.Function(
	"filepath_dir",
	func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, NewInvalidNumberOfArgumentsError("filepath_dir", "", 1, len(params))
		}

		val, ok := params[0].(string)
		if !ok {
			return nil, NewInvalidArgumentTypeError("filepath_dir", "input must be string")
		}

		return filepath.Dir(val), nil
	},
	filepath.Dir,
)

func UniqSlice[T cmp.Ordered](in []T) []T {
	slices.Sort(in)

	return slices.Compact(in)
}

// Uniq takes a list of strings or interface{}, sorts them
// and remove duplicated values
var Uniq = expr.Function(
	"uniq",
	func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, NewInvalidNumberOfArgumentsError("uniq", "", 1, len(args))
		}

		arg := args[0]

		switch val := arg.(type) {
		case []any:
			var result []string

			for _, v := range val {
				result = append(result, fmt.Sprintf("%s", v))
			}

			return UniqSlice(result), nil

		case []string:
			return UniqSlice(val), nil

		default:
			return nil, NewInvalidArgumentTypeError("uniq", fmt.Sprintf("invalid input, must be an array of [string] or [interface], got %T", arg))
		}
	},
	new(func([]string) []string),
	new(func([]any) []string),
)

// Override built-in duration() function to provide support for additional periods
// - 'd' (day)
// - 'w' (week)
// - 'm' (month)
var Duration = expr.Function(
	"duration",
	func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, NewInvalidNumberOfArgumentsError("duration", "", 1, len(args))
		}

		val, ok := args[0].(string)
		if !ok {
			return nil, NewInvalidArgumentTypeError("duration", fmt.Sprintf("invalid input, must be a string, got %T", args[0]))
		}

		return str2duration.ParseDuration(val)
	},
	time.ParseDuration,
)

var LimitPathDepthTo = expr.Function(
	"limit_path_depth_to",
	func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, NewInvalidNumberOfArgumentsError("limit_path_depth_to", "", 2, len(args))
		}

		input, ok := args[0].(string)
		if !ok {
			return nil, NewInvalidArgumentTypeError("limit_path_depth_to", fmt.Sprintf("invalid input, first argument must be a 'string', got %T", args[0]))
		}

		length, ok := args[1].(int)
		if !ok {
			return nil, NewInvalidArgumentTypeError("limit_path_depth_to", fmt.Sprintf("invalid input, first argument must be an 'int', got %T", args[0]))
		}

		chunks := strings.Split(input, "/")
		if len(chunks) <= length {
			return input, nil
		}

		return strings.Join(chunks[0:length-1], "/"), nil // nosemgrep
	},
	new(func(string, int) string),
)
