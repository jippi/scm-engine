package scm

import (
	"context"
	"io"
)

type Client interface {
	Labels() LabelClient
	MergeRequests() MergeRequestClient
	EvalContext(ctx context.Context) (EvalContext, error)
	ApplyStep(ctx context.Context, update *UpdateMergeRequestOptions, step EvaluationActionStep) error
}

type LabelClient interface {
	List(ctx context.Context) ([]*Label, error)
	Create(ctx context.Context, opt *CreateLabelOptions) (*Label, *Response, error)
	Update(ctx context.Context, opt *UpdateLabelOptions) (*Label, *Response, error)
}

type MergeRequestClient interface {
	Update(ctx context.Context, opt *UpdateMergeRequestOptions) (*Response, error)
	List(ctx context.Context, options *ListMergeRequestsOptions) ([]ListMergeRequest, error)
	GetRemoteConfig(ctx context.Context, name string, ref string) (io.Reader, error)
}

type EvalContext interface {
	IsValid() bool
	SetWebhookEvent(in any)
}

type EvalContextualizer struct{}

func (e EvalContextualizer) _isEvalContext() {}
