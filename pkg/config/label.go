package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/patcher"
	"github.com/expr-lang/expr/vm"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/stdlib"
	"github.com/jippi/scm-engine/pkg/tui"
	"github.com/jippi/scm-engine/pkg/types"
	slogctx "github.com/veqryn/slog-context"
)

// labelType is a custom type for our enum
type labelType string

const (
	ConditionalLabel labelType = "conditional"
	GenerateLabels   labelType = "generate"
)

type Labels []*Label

func (labels Labels) Evaluate(ctx context.Context, evalContext scm.EvalContext) ([]scm.EvaluationResult, error) {
	var results []scm.EvaluationResult

	// Evaluate labels
	for _, label := range labels {
		ctx := slogctx.With(ctx, slog.String("label_name", label.Name))

		slogctx.Debug(ctx, "Evaluating label")

		evaluationResult, err := label.Evaluate(ctx, evalContext)
		if err != nil {
			return nil, fmt.Errorf("label: %s; %w", label.Name, err)
		}

		if evaluationResult == nil {
			slogctx.Debug(ctx, "Label evaluated negatively, skipping")

			continue
		}

		slogctx.Debug(ctx, "Label evaluation done", slog.Any("label_eval_result", evaluationResult))

		results = append(results, evaluationResult...)
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

type Label struct {
	// (Optional) Strategy used for the label
	//
	// - "conditional" will, based on the boolean output of [script], control if the label [name] should be added/removed on the MR
	// - "computed" will, based on the list of strings output of [script], add/remove labels on the MR
	//
	// See: https://jippi.github.io/scm-engine/configuration/#label.strategy
	Strategy labelType `json:"strategy,omitempty" yaml:"strategy,omitempty" jsonschema:"default=conditional,enum=conditional,enum=generate"`

	// Name of the label being generated.
	//
	// May only be used with [conditional] labelling type
	//
	// See: https://jippi.github.io/scm-engine/configuration/#label.name
	Name string `json:"name,omitempty" yaml:"name,omitempty" jsonschema:"dependentRequired"`

	// (Optional) Description for the label being generated.
	//
	// See: https://jippi.github.io/scm-engine/configuration/#label.description
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// (Optional) The HEX color code to use for the label.
	//
	// May use the color variables (e.g., $purple-300) defined in Twitter Bootstrap
	// https://getbootstrap.com/docs/5.3/customize/color/#all-colors
	//
	// See: https://jippi.github.io/scm-engine/configuration/#label.color
	Color string `json:"color,omitempty" yaml:"color,omitempty"`

	// (Optional) Priority controls wether the label should be a priority label or regular one.
	//
	// This controls if the label is prioritized (sorted first) in the list.
	//
	// See: https://jippi.github.io/scm-engine/configuration/#label.priority
	Priority types.Value[int] `json:"priority,omitempty" yaml:"priority,omitempty"`

	// Script contains the (https://expr-lang.org/) script used to emit labels for the MR.
	//
	// See: https://jippi.github.io/scm-engine/configuration/#label.script
	Script string `json:"script" yaml:"script,omitempty"`

	// SkipIf is an optional (https://expr-lang.org/) script, returning a boolean, wether to
	// skip (true) or process (false) this label step.
	//
	// See: https://jippi.github.io/scm-engine/configuration/#label.skip_if
	SkipIf string `json:"skip_if,omitempty" yaml:"skip_if,omitempty"`

	//
	// -- Internal state
	//

	// scriptCompiled is the [expr-lang](https://expr-lang.org/) [Script] script pre-compiled
	scriptCompiled *vm.Program `json:"-" yaml:"-"`

	// skipIfCompiled is the [expr-lang](https://expr-lang.org/) [SkipIf] script pre-compiled
	skipIfCompiled *vm.Program `json:"-" yaml:"-"`

	expectedReturnType any `json:"-" yaml:"-"`
}

func (p *Label) Setup(evalContext scm.EvalContext) error {
	var scriptReturnType expr.Option

	if len(p.Script) == 0 {
		return errors.New("required 'script' field is empty")
	}

	// Default behavior is conditional labels
	if p.Strategy == "" {
		p.Strategy = ConditionalLabel
	}

	// Validation and label type specific initialization

	switch p.Strategy {
	case GenerateLabels:
		if p.Name != "" {
			return fmt.Errorf("[name] may only be specified when using [type: %q]", ConditionalLabel)
		}

		p.expectedReturnType = []string{}
		scriptReturnType = expr.AsKind(reflect.TypeFor[[]string]().Kind())

	case ConditionalLabel:
		if p.Name == "" {
			return fmt.Errorf("[name] is required when using [type: %q]", ConditionalLabel)
		}

		p.expectedReturnType = true
		scriptReturnType = expr.AsBool()

	default:
		return fmt.Errorf("unknown label [type] %q. use %q or %q", p.Strategy, GenerateLabels, ConditionalLabel)
	}

	var err error

	if p.scriptCompiled == nil {
		p.Color = tui.Replace(p.Color)

		opts := []expr.Option{}
		opts = append(opts, scriptReturnType)
		opts = append(opts, expr.Env(evalContext))
		opts = append(opts, stdlib.FunctionRenamer)
		opts = append(opts, stdlib.Functions...)
		opts = append(opts, expr.Patch(patcher.WithContext{Name: "ctx"}))

		p.scriptCompiled, err = expr.Compile(p.Script, opts...)
		if err != nil {
			return fmt.Errorf("could not compile 'script' into valid expr-lang syntax: %w", err)
		}
	}

	if p.skipIfCompiled == nil && len(p.SkipIf) > 0 {
		p.Color = tui.Replace(p.Color)

		opts := []expr.Option{}
		opts = append(opts, expr.AsBool())
		opts = append(opts, expr.Env(evalContext))
		opts = append(opts, stdlib.FunctionRenamer)
		opts = append(opts, stdlib.Functions...)
		opts = append(opts, expr.Patch(patcher.WithContext{Name: "ctx"}))

		p.skipIfCompiled, err = expr.Compile(p.SkipIf, opts...)
		if err != nil {
			return fmt.Errorf("could not compile 'if' into valid expr-lang syntax: %w", err)
		}
	}

	return nil
}

func (p *Label) ShouldSkip(ctx context.Context, evalContext scm.EvalContext) (bool, error) {
	if err := p.Setup(evalContext); err != nil {
		return true, err
	}

	return runAndCheckBool(ctx, p.skipIfCompiled, evalContext)
}

func (p *Label) Evaluate(ctx context.Context, evalContext scm.EvalContext) ([]scm.EvaluationResult, error) {
	if err := p.Setup(evalContext); err != nil {
		return nil, fmt.Errorf("failed to initialize expr script engine: %w", err)
	}

	// Check if the label should be skipped
	if skip, err := p.ShouldSkip(ctx, evalContext); err != nil || skip {
		return nil, err
	}

	// Run the compiled expr-lang script
	output, err := expr.Run(p.scriptCompiled, evalContext)
	if err != nil {
		return nil, err
	}

	var result []scm.EvaluationResult

	switch outputValue := output.(type) {
	case bool:
		if p.Strategy != ConditionalLabel {
			return nil, errors.New("Script returned an unexpected boolean; Did you forget the 'type: computed' on your label?")
		}

		result = append(result, p.resultForLabel(p.Name, outputValue))

	// When using 'uniq' function, the result is a correct []string slice
	case []string:
		if p.Strategy != GenerateLabels {
			return nil, errors.New("Script returned an unexpected list of strings; Did you forget the 'type: computed' on your label?")
		}

		for _, label := range outputValue {
			result = append(result, p.resultForLabel(label, true))
		}

	// In some cases the slice can be of 'any' type, thats fine, as long as the underlying type is 'string'
	case []any:
		if p.Strategy != GenerateLabels {
			return nil, errors.New("Script returned an unexpected list of strings; Did you forget the 'type: computed' on your label?")
		}

		for _, label := range outputValue {
			switch labelVal := label.(type) {
			case string:
				result = append(result, p.resultForLabel(labelVal, true))

			default:
				return nil, fmt.Errorf("Script must return a list of strings but encountered a value of type %T (%v)", labelVal, labelVal)
			}
		}

	default:
		return nil, fmt.Errorf("rule evaluation returned %T (%+v); must return %T", output, output, p.expectedReturnType)
	}

	return result, nil
}

func (p Label) resultForLabel(name string, matched bool) scm.EvaluationResult {
	return scm.EvaluationResult{
		Name:        name,
		Matched:     matched,
		Color:       p.Color,
		Description: p.Description,
		Priority:    p.Priority,
	}
}

func runAndCheckBool(ctx context.Context, program *vm.Program, evalContext scm.EvalContext) (bool, error) {
	if program == nil {
		return false, nil
	}

	output, err := expr.Run(program, evalContext)
	if err != nil {
		return false, err
	}

	switch outputValue := output.(type) {
	case bool:
		slogctx.Debug(ctx, "script eval done", slog.Bool("script_outcome", outputValue))

		return outputValue, nil

	default:
		return false, fmt.Errorf("rule evaluation returned %T (%v); must return %T", output, output, true)
	}
}
