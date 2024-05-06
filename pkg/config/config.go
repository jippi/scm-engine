package config

import (
	"errors"
	"fmt"

	"github.com/jippi/scm-engine/pkg/scm"
)

type Config struct {
	Labels Labels `yaml:"label"`
}

func (c Config) Evaluate(e scm.EvalContext) ([]scm.EvaluationResult, error) {
	results, err := c.Labels.Evaluate(e)
	if err != nil {
		return nil, err
	}

	// Sanity/validation checks
	seen := map[string]bool{}

	for _, result := range results {
		// Check labels has a proper name
		if len(result.Name) == 0 {
			return nil, errors.New("A label was generated with empty name, please check your configuration.")
		}

		// Check uniqueness of labels
		if _, ok := seen[result.Name]; ok {
			return nil, fmt.Errorf("The label %q was generated multiple times, please check your configuration. Hint: If you use [compute] label type, you can use the 'uniq()' function (example: '| uniq()')", result.Name)
		}

		seen[result.Name] = true
	}

	return results, nil
}
