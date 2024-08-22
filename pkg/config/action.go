package config

import (
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/patcher"
	"github.com/expr-lang/expr/vm"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/stdlib"
)

type Actions []Action

func (actions Actions) Evaluate(evalContext scm.EvalContext) ([]Action, error) {
	results := []Action{}

	// Evaluate actions
	for _, action := range actions {
		ok, err := action.Evaluate(evalContext)
		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		results = append(results, action)
	}

	return results, nil
}

type Action scm.EvaluationActionResult

func (p *Action) Evaluate(evalContext scm.EvalContext) (bool, error) {
	program, err := p.initialize(evalContext)
	if err != nil {
		return false, err
	}

	// Run the compiled expr-lang script
	return runAndCheckBool(program, evalContext)
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
