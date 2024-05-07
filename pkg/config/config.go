package config

import (
	"github.com/jippi/scm-engine/pkg/scm"
)

type Config struct {
	Labels  Labels  `yaml:"label"`
	Actions Actions `yaml:"actions"`
}

func (c Config) Evaluate(evalContext scm.EvalContext) ([]scm.EvaluationLabelResult, []Action, error) {
	labels, err := c.Labels.Evaluate(evalContext)
	if err != nil {
		return nil, nil, err
	}

	actions, err := c.Actions.Evaluate(evalContext)
	if err != nil {
		return nil, nil, err
	}

	return labels, actions, nil
}
