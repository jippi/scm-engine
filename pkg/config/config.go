package config

import (
	"fmt"

	"github.com/jippi/gitlab-labeller/pkg/scm"
)

type Config struct {
	Labels Labels `yaml:"label"`
}

func (c Config) Evaluate(e scm.EvalContext) ([]scm.EvaluationResult, error) {
	res, err := c.Labels.Evaluate(e)
	if err != nil {
		return nil, err
	}

	// Sanity/validation checks
	seen := map[string]bool{}
	for _, r := range res {
		// Check labels has a proper name
		if len(r.Name) == 0 {
			return nil, fmt.Errorf("A label was generated with empty name, please check your configuration.")
		}

		// Check uniqueness of labels
		if _, ok := seen[r.Name]; ok {
			return nil, fmt.Errorf("The label %q was generated multiple times, please check your configuration. Hint: If you use [compute] label type, you can use the 'uniq()' function (example: '| uniq()')", r.Name)
		}

		seen[r.Name] = true
	}

	return res, nil
}
