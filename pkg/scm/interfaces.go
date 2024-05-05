package scm

import (
	"context"
)

type Client interface {
	Labels() LabelClient
	MergeRequests() MergeRequestClient
	EvalContext(ctx context.Context) (EvalContext, error)
}

type LabelClient interface {
	List(ctx context.Context) ([]*Label, error)
	Create(ctx context.Context, opt *CreateLabelOptions) (*Label, *Response, error)
	Update(ctx context.Context, opt *UpdateLabelOptions) (*Label, *Response, error)
}

type MergeRequestClient interface {
	Update(ctx context.Context, opt *UpdateMergeRequestOptions) (*Response, error)
}

type EvalContext interface {
	_isEvalContext()
}

type EvalContextualizer struct{}

func (e EvalContextualizer) _isEvalContext() {}
