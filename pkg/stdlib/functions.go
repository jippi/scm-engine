package stdlib

import (
	"cmp"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/xhit/go-str2duration/v2"
)

var FilepathDir = expr.Function(
	"filepath_dir",
	func(params ...any) (any, error) {
		if len(params) != 1 {
			return nil, fmt.Errorf("filepath_dir: accepts exactly 1 argument, %d provided", len(params))
		}

		val, ok := params[0].(string)
		if !ok {
			return nil, errors.New("input to filepath_dir must be of type 'string'")
		}

		return filepath.Dir(val), nil
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

		switch val.Kind() { //nolint:exhaustive
		case reflect.Slice:
			switch val.Type().Elem().Kind() { //nolint:exhaustive
			case reflect.Interface:
				var x []string
				for _, v := range arg.([]any) { //nolint:forcetypeassert
					x = append(x, fmt.Sprintf("%s", v))
				}

				return UniqSlice(x), nil

			case reflect.String:
				return UniqSlice(arg.([]string)), nil //nolint:forcetypeassert

			}
		}

		return nil, errors.New("invalid type")
	},
)

var Duration = expr.Function(
	"duration",
	func(args ...any) (any, error) {
		val, ok := args[0].(string)
		if !ok {
			return nil, errors.New("input to duration() must be of type 'string'")
		}

		return str2duration.ParseDuration(val)
	},
	str2duration.ParseDuration,
)

var LimitPathDepthTo = expr.Function(
	"limit_path_depth_to",
	func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("limit_path_depth_to() expect exactly two arguments")
		}

		input, ok := args[0].(string)
		if !ok {
			return nil, errors.New("first input to limit_path_depth_to() must be of type 'string'")
		}

		length, ok := args[1].(int)
		if !ok {
			return nil, errors.New("second input to limit_path_depth_to() must be of type 'int'")
		}

		chunks := strings.Split(input, "/")
		if len(chunks) <= length {
			return input, nil
		}

		return strings.Join(chunks[0:length-1], "/"), nil
	},
)
