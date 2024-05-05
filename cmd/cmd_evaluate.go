package cmd

import (
	"fmt"

	"github.com/jippi/gitlab-labeller/pkg/state"
	"github.com/urfave/cli/v2"
)

func Evaluate(cCtx *cli.Context) error {
	ctx := state.ContextWithProjectID(cCtx.Context, cCtx.String(FlagSCMProject))

	switch {
	// If the flag is set, use that for evaluation
	case cCtx.String(FlagMergeRequestID) != "":
		return ProcessMR(ctx, cCtx, cCtx.String(FlagMergeRequestID))

	// If no flag is set, we require arguments
	case cCtx.Args().Len() == 0:
		return fmt.Errorf("Missing required argument: %s", FlagMergeRequestID)

	default:
		for _, mr := range cCtx.Args().Slice() {
			if err := ProcessMR(ctx, cCtx, mr); err != nil {
				return err
			}
		}

		return nil
	}
}
