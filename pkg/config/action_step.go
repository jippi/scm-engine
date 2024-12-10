package config

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type actionList struct {
	name     string
	instance any
}

var actions = []actionList{
	{name: "add_label", instance: AddLabelAction{}},
	{name: "approve", instance: ApproveAction{}},
	{name: "assign_reviewers", instance: AssignReviewers{}},
	{name: "close", instance: CloseAction{}},
	{name: "comment", instance: CommentAction{}},
	{name: "lock_discussion", instance: LockDiscussionAction{}},
	{name: "remove_label", instance: RemoveLabelAction{}},
	{name: "reopen", instance: ReopenAction{}},
	{name: "unapprove", instance: UnapproveAction{}},
	{name: "unlock_discussion", instance: UnlockDiscussionAction{}},
	{name: "update_description", instance: UpdateDescriptionAction{}},
}

type BaseAction struct {
	// The action to take
	//
	// See: https://jippi.github.io/scm-engine/configuration/#actions.if.then.action
	Action string `json:"action" yaml:"action"`
}

// Hello World?
type ApproveAction struct {
	BaseAction
}

type UnapproveAction struct {
	BaseAction
}

type LockDiscussionAction struct {
	BaseAction
}

type CloseAction struct {
	BaseAction
}

type ReopenAction struct {
	BaseAction
}

type RemoveLabelAction struct {
	BaseAction

	// The label name to remove.
	//
	// See: https://jippi.github.io/scm-engine/configuration/#actions.if.then.action
	Label string `json:"label" yaml:"label"`
}

type CommentAction struct {
	BaseAction

	// The message that will be commented on the Merge Request
	//
	// See: https://jippi.github.io/scm-engine/configuration/#actions.if.then.action
	Message string `json:"message" yaml:"message"`
}

type AssignReviewers struct {
	BaseAction

	// The source of the reviewers
	Source *string `json:"source,omitempty" yaml:"source,omitempty" jsonschema:"enum=codeowners"`
	// The max number of reviewers to assign
	Limit int `json:"limit,omitempty" yaml:"limit,omitempty"`
	// The mode of assigning reviewers
	Mode string `json:"mode,omitempty" yaml:"mode,omitempty" jsonschema:"enum=random"`
}

type AddLabelAction struct {
	BaseAction

	// The label name to add.
	//
	// See: https://jippi.github.io/scm-engine/configuration/#actions.if.then.action
	Label string `json:"label" yaml:"label"`
}

type UnlockDiscussionAction struct {
	BaseAction
}

// Updates the Merge Request Description
type UpdateDescriptionAction struct {
	BaseAction

	// A list of key/value pairs to replace in the description.
	// The key is the raw string to replace in the Merge Request description.
	// The value is an Expr Lang expression returning a string that key will be replaced with
	//
	// See: https://jippi.github.io/scm-engine/configuration/#actions.if.then.action
	Replace map[string]string `json:"replace" yaml:"replace"`
}

// This key controls what kind of action that should be taken.
//
// See: https://jippi.github.io/scm-engine/configuration/#actions.if.then
type ActionStep map[string]any

func (step ActionStep) JSONSchema() *jsonschema.Schema {
	configs := []*jsonschema.Schema{}
	validActions := []any{}

	r := new(jsonschema.Reflector)
	if err := r.AddGoComments("github.com/jippi/scm-engine", "./"); err != nil {
		panic(err)
	}

	for _, action := range actions {
		reflected := r.Reflect(action.instance)
		definitionID := strings.ReplaceAll(reflected.Ref, "#/$defs/", "")

		actionSchema := &jsonschema.Schema{
			If: &jsonschema.Schema{
				Properties: orderedmap.New[string, *jsonschema.Schema](
					orderedmap.WithInitialData(
						orderedmap.Pair[string, *jsonschema.Schema]{
							Key:   "action",
							Value: &jsonschema.Schema{Const: action.name},
						},
					),
				),
			},
			Then: reflected.Definitions[definitionID],
		}

		configs = append(configs, actionSchema)
		validActions = append(validActions, action.name)
	}

	// https://json-schema.org/understanding-json-schema/reference/conditionals#ifthenelse
	return &jsonschema.Schema{
		Type:     "object",
		Required: []string{"action"},
		Properties: orderedmap.New[string, *jsonschema.Schema](
			orderedmap.WithInitialData(
				orderedmap.Pair[string, *jsonschema.Schema]{
					Key:   "action",
					Value: &jsonschema.Schema{Type: "string", Enum: validActions},
				},
			),
		),
		AllOf: configs,
	}
}

func (step ActionStep) RequiredInt(name string) (int, error) {
	value, ok := step[name]
	if !ok {
		return 0, fmt.Errorf("Required 'step' key '%s' is missing", name)
	}

	valueInt, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("Required 'step' key '%s' must be of type int, got %T", name, value)
	}

	return valueInt, nil
}

func (step ActionStep) RequiredString(name string) (string, error) {
	value, ok := step[name]
	if !ok {
		return "", fmt.Errorf("Required 'step' key '%s' is missing", name)
	}

	valueString, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("Required 'step' key '%s' must be of type string, got %T", name, value)
	}

	return valueString, nil
}

func (step ActionStep) RequiredStringEnum(name string, values ...string) (string, error) {
	value, ok := step[name]
	if !ok {
		return "", fmt.Errorf("Required 'step' key '%s' is missing", name)
	}

	valueString, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("Required 'step' key '%s' must be of type string, got %T", name, value)
	}

	for _, validValue := range values {
		if valueString == validValue {
			return valueString, nil
		}
	}

	return "", fmt.Errorf("Required 'step' key '%s' must be one of %v, got %s", name, values, valueString)
}

func (step ActionStep) OptionalInt(name string, fallback int) (int, error) {
	value, ok := step[name]
	if !ok {
		return fallback, nil
	}

	valueInt, ok := value.(int)
	if !ok {
		return fallback, fmt.Errorf("Optional step field '%s' must be of type int, got %T", name, value)
	}

	return valueInt, nil
}

func (step ActionStep) OptionalString(name, defaultValue string) (string, error) {
	value, ok := step[name]
	if !ok {
		return defaultValue, nil
	}

	valueString, ok := value.(string)
	if !ok {
		return defaultValue, fmt.Errorf("Optional step field '%s' must be of type string, got %T", name, value)
	}

	return valueString, nil
}

func (step ActionStep) OptionalStringEnum(name string, fallback string, values ...string) (string, error) {
	value, ok := step[name]
	if !ok {
		return fallback, nil
	}

	valueString, ok := value.(string)
	if !ok {
		return fallback, fmt.Errorf("Optional step field '%s' must be of type string, got %T", name, value)
	}

	for _, validValue := range values {
		if valueString == validValue {
			return valueString, nil
		}
	}

	return fallback, fmt.Errorf("Optional step field '%s' must be one of %v, got %s", name, values, valueString)
}

func (step ActionStep) Get(name string) (any, error) {
	value, ok := step[name]
	if !ok {
		return nil, fmt.Errorf("'step' key '%s' is missing", name)
	}

	return value, nil
}
