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
		return filepath.Dir(params[0].(string)), nil //nolint:forcetypeassert
	},
	filepath.Dir, // string => string
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
		switch elements := args[0].(type) {
		case []any:
			var result []string

			for _, element := range elements {
				result = append(result, fmt.Sprintf("%s", element))
			}

			return UniqSlice(result), nil

		case []string:
			return UniqSlice(elements), nil

		default:
			return nil, fmt.Errorf("invalid input, must be an array of [string] or [interface], got %T", args[0])
		}
	},
	new(func([]any) []string),    // []any -> []string (when using map() that always return []any)
	new(func([]string) []string), // []string -> []string
)

// Override built-in duration() function to provide support for additional periods
// - 'd' (day)
// - 'w' (week)
// - 'm' (month)
var Duration = expr.Function(
	"duration",
	func(args ...any) (any, error) {
		return str2duration.ParseDuration(args[0].(string)) //nolint:forcetypeassert
	},
	time.ParseDuration, // string => (time.Duration, error)
)

var LimitPathDepthTo = expr.Function(
	"limit_path_depth_to",
	func(args ...any) (any, error) {
		input := args[0].(string) //nolint:forcetypeassert
		length := args[1].(int)   //nolint:forcetypeassert

		chunks := strings.Split(input, "/")
		if len(chunks) <= length {
			return input, nil
		}

		return strings.Join(chunks[0:length-1], "/"), nil // nosemgrep
	},
	new(func(string, int) string), // (string, int) => string
)
