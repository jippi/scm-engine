package cmd

import (
	"fmt"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/urfave/cli/v2"
)

func Evaluate(cCtx *cli.Context) error {
	ctx := state.WithProjectID(cCtx.Context, cCtx.String(FlagSCMProject))
	ctx = state.WithCommitSHA(ctx, cCtx.String(FlagCommitSHA))
	ctx = state.WithToken(ctx, cCtx.String(FlagAPIToken))
	ctx = state.WithUpdatePipeline(ctx, cCtx.Bool(FlagUpdatePipeline), cCtx.String(FlagUpdatePipelineURL))

	cfg, err := config.LoadFile(cCtx.String(FlagConfigFile))
	if err != nil {
		return err
	}

	client, err := getClient(ctx)
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
			ctx := state.WithMergeRequestID(ctx, mr.ID)
			ctx = state.WithCommitSHA(ctx, mr.SHA)

			if err := ProcessMR(ctx, client, cfg, nil); err != nil {
				return err
			}
		}

	// If the flag is set, use that for evaluation
	case cCtx.String(FlagMergeRequestID) != "":
		ctx = state.WithMergeRequestID(ctx, cCtx.String(FlagMergeRequestID))

		return ProcessMR(ctx, client, cfg, nil)

	// If no flag is set, we require arguments
	case cCtx.Args().Len() == 0:
		return fmt.Errorf("Missing required argument: %s", FlagMergeRequestID)

	default:
		for _, mr := range cCtx.Args().Slice() {
			ctx = state.WithMergeRequestID(ctx, mr)

			if err := ProcessMR(ctx, client, cfg, nil); err != nil {
				return err
			}
		}
	}

	return nil
}
