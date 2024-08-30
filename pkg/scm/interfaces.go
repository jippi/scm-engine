package scm

import (
	"context"
	"io"
)

type Client interface {
	ApplyStep(ctx context.Context, evalContext EvalContext, update *UpdateMergeRequestOptions, step EvaluationActionStep) error
	EvalContext(ctx context.Context) (EvalContext, error)
	FindMergeRequestsForPeriodicEvaluation(ctx context.Context, filters MergeRequestListFilters) ([]PeriodicEvaluationMergeRequest, error)
	Labels() LabelClient
	MergeRequests() MergeRequestClient
	Start(ctx context.Context) error
	Stop(ctx context.Context, err error) error
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
	CanUseConfigurationFileFromChangeRequest(ctx context.Context) bool
	GetDescription() string
	IsValid() bool
	SetContext(ctx context.Context)
	SetWebhookEvent(in any)
	TrackActionGroupExecution(name string)
	HasExecutedActionGroup(name string) bool
}
