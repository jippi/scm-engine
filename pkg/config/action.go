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

type (
	Actions []Action

	Action struct {
		// The name of the action, this is purely for debugging and your convenience.
		//
		// See: https://jippi.github.io/scm-engine/configuration/#actions.name
		Name string `json:"name" yaml:"name"`

		// (Optional) Only one action per group (in order) will be executed per evaluation cycle.
		// Use this to 'stop' other actions from running with the same group name
		Group string `json:"group,omitempty" yaml:"group"`

		// A key controlling if the action should executed or not.
		//
		// This script is in Expr-lang: https://expr-lang.org/docs/language-definition
		//
		// See: https://jippi.github.io/scm-engine/configuration/#actions.if
		If string `json:"if" yaml:"if"`

		// The list of operations to take if the action.if returned true.
		//
		// See: https://jippi.github.io/scm-engine/configuration/#actions.if.then
		Then []ActionStep `json:"then" yaml:"then"`
	}
)

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

func (p *Action) Evaluate(ctx context.Context, evalContext scm.EvalContext) (bool, error) {
	program, err := p.Setup(evalContext)
	if err != nil {
		return false, err
	}

	// Run the compiled expr-lang script
	return runAndCheckBool(ctx, program, evalContext)
}

func (p *Action) Setup(evalContext scm.EvalContext) (*vm.Program, error) {
	opts := []expr.Option{}
	opts = append(opts, expr.AsBool())
	opts = append(opts, expr.Env(evalContext))
	opts = append(opts, stdlib.FunctionRenamer)
	opts = append(opts, stdlib.Functions...)
	opts = append(opts, expr.Patch(patcher.WithContext{Name: "ctx"}))

	return expr.Compile(p.If, opts...)
}
