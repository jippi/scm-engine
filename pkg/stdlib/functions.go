package stdlib

import (
	"cmp"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/xhit/go-str2duration/v2"
)

var FilepathDir = expr.Function(
	"filepath_dir",
	func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("filepath_dir: accepts exactly 1 argument, %d provided", len(params))
		}

		return filepath.Dir(params[0].(string)), nil
	},
	filepath.Dir,
)

func UniqSlice[T cmp.Ordered](in []T) []T {
	slices.Sort(in)

	return slices.Compact(in)
}

var Uniq = expr.Function(
	"uniq",
	func(params ...any) (any, error) {
		arg := params[0]
		val := reflect.ValueOf(arg)

		switch val.Kind() {
		case reflect.Slice:
			switch val.Type().Elem().Kind() {
			case reflect.Interface:
				var x []string
				for _, v := range arg.([]any) {
					x = append(x, fmt.Sprintf("%s", v))
				}

				return UniqSlice(x), nil

			case reflect.String:
				return UniqSlice(arg.([]string)), nil
			}
		}

		return nil, fmt.Errorf("invalid type")
	},
)

var Duration = expr.Function(
	"duration",
	func(args ...any) (any, error) {
		return str2duration.ParseDuration(args[0].(string))
	},
	str2duration.ParseDuration,
)
