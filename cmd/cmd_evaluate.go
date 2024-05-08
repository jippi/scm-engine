package cmd

import (
	"fmt"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/urfave/cli/v2"
)

func Evaluate(cCtx *cli.Context) error {
	ctx := state.ContextWithProjectID(cCtx.Context, cCtx.String(FlagSCMProject))

	cfg, err := config.LoadFile(cCtx.String(FlagConfigFile))
	if err != nil {
		return err
	}

	client, err := gitlab.NewClient(cCtx.String(FlagAPIToken), cCtx.String(FlagSCMBaseURL))
	if err != nil {
		return err
	}

	switch {
	// If first arg is 'all' we will find all opened MRs and apply the rules to them
	case cCtx.Args().First() == "all":
		res, err := client.MergeRequests().List(ctx, &scm.ListMergeRequestsOptions{State: "opened", First: 100})
		if err != nil {
			return err
		}

		for _, mr := range res {
			if err := ProcessMR(ctx, client, cfg, mr.ID); err != nil {
				return err
			}
		}

	// If the flag is set, use that for evaluation
	case cCtx.String(FlagMergeRequestID) != "":
		return ProcessMR(ctx, client, cfg, cCtx.String(FlagMergeRequestID))

	// If no flag is set, we require arguments
	case cCtx.Args().Len() == 0:
		return fmt.Errorf("Missing required argument: %s", FlagMergeRequestID)

	default:
		for _, mr := range cCtx.Args().Slice() {
			if err := ProcessMR(ctx, client, cfg, mr); err != nil {
				return err
			}
		}
	}

	return nil
}
