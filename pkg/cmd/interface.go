package cmd

import (
	"context"

	"github.com/urfave/cli/v2"
)

type Commands interface {
	Evaluate(ctx context.Context, cCtx *cli.Context, mr string)
}
