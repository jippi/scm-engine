package scm

import (
	"context"
	"io"
)

type Client interface {
	ApplyStep(ctx context.Context, evalContext EvalContext, update *UpdateMergeRequestOptions, step ActionStep) error
	EvalContext(ctx context.Context) (EvalContext, error)
	FindMergeRequestsForPeriodicEvaluation(ctx context.Context, filters MergeRequestListFilters) ([]PeriodicEvaluationMergeRequest, error)
	GetProjectFiles(ctx context.Context, project string, ref *string, files []string) (map[string]string, error)
	Labels() LabelClient
	MergeRequests() MergeRequestClient
	Start(ctx context.Context) error
	Stop(ctx context.Context, err error, allowPipelineFailure bool) error
}

type LabelClient interface {
	Create(ctx context.Context, opt *CreateLabelOptions) (*Label, *Response, error)
	List(ctx context.Context) ([]*Label, error)
	Update(ctx context.Context, opt *UpdateLabelOptions) (*Label, *Response, error)
}

type MergeRequestClient interface {
	GetRemoteConfig(ctx context.Context, name string, ref string) (io.Reader, error)
	List(ctx context.Context, options *ListMergeRequestsOptions) ([]ListMergeRequest, error)
	Update(ctx context.Context, opt *UpdateMergeRequestOptions) (*Response, error)
}

type EvalContext interface {
	AllowPipelineFailure(ctx context.Context) bool
	CanUseConfigurationFileFromChangeRequest(ctx context.Context) bool
	GetDescription() string
	HasExecutedActionGroup(name string) bool
	IsValid() bool
	SetContext(ctx context.Context)
	SetWebhookEvent(in any)
	TrackActionGroupExecution(name string)
}

type ActionStep interface {
	RequiredString(name string) (string, error)
	RequiredInt(name string) (int, error)
	OptionalString(name, fallback string) (string, error)
	Get(name string) (any, error)
}
