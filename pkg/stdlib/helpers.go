package stdlib

import (
	"fmt"
	"time"

	"github.com/xhit/go-str2duration/v2"
)

func ToDuration(input any) time.Duration {
	switch val := input.(type) {
	case time.Duration:
		return val

	case string:
		dur, err := str2duration.ParseDuration(val)
		if err != nil {
			panic(err)
		}

		return dur

	default:
		panic(fmt.Errorf("unsupported input type for duration: %T", val))
	}
}
