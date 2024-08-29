package config

import (
	"context"
	"log/slog"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/patcher"
	"github.com/expr-lang/expr/vm"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/stdlib"
	slogctx "github.com/veqryn/slog-context"
)

type Actions []Action

func (actions Actions) Evaluate(ctx context.Context, evalContext scm.EvalContext) ([]Action, error) {
	results := []Action{}

	// Evaluate actions
	for _, action := range actions {
		ctx := slogctx.With(ctx, slog.String("action_name", action.Name))

		slogctx.Debug(ctx, "Evaluating action")

		ok, err := action.Evaluate(ctx, evalContext)
		if err != nil {
			return nil, err
		}

		if !ok {
			slogctx.Debug(ctx, "Action evaluated negatively, skipping")

			continue
		}

		slogctx.Debug(ctx, "Action evaluated positively")

		results = append(results, action)
	}

	return results, nil
}

type Action scm.EvaluationActionResult

func (p *Action) Evaluate(ctx context.Context, evalContext scm.EvalContext) (bool, error) {
	program, err := p.initialize(evalContext)
	if err != nil {
		return false, err
	}

	// Run the compiled expr-lang script
	return runAndCheckBool(ctx, program, evalContext)
}

func (p *Action) initialize(evalContext scm.EvalContext) (*vm.Program, error) {
	opts := []expr.Option{}
	opts = append(opts, expr.AsBool())
	opts = append(opts, expr.Env(evalContext))
	opts = append(opts, stdlib.FunctionRenamer)
	opts = append(opts, stdlib.Functions...)
	opts = append(opts, expr.Patch(patcher.WithContext{Name: "ctx"}))

	return expr.Compile(p.If, opts...)
}
