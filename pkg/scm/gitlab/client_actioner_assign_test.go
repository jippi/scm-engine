package gitlab_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type evalContextMock struct {
	mock.Mock
}

func (c *evalContextMock) IsValid() bool {
	return c != nil
}

func (c *evalContextMock) SetWebhookEvent(in any) {
	c.Called(in)
}

func (c *evalContextMock) SetContext(ctx context.Context) {
	c.Called(ctx)
}

func (c *evalContextMock) GetDescription() string {
	args := c.Called()

	return args.String(0)
}

func (c *evalContextMock) CanUseConfigurationFileFromChangeRequest(ctx context.Context) bool {
	args := c.Called(ctx)

	return args.Bool(0)
}

func (c *evalContextMock) AllowPipelineFailure(ctx context.Context) bool {
	args := c.Called(ctx)

	return args.Bool(0)
}

func (c *evalContextMock) TrackActionGroupExecution(group string) {
	c.Called(group)
}

func (c *evalContextMock) HasExecutedActionGroup(group string) bool {
	args := c.Called(group)

	return args.Bool(0)
}

func (c *evalContextMock) GetCodeOwners() scm.Actors {
	args := c.Called()

	if actors, ok := args.Get(0).(scm.Actors); ok {
		return actors
	}

	return nil
}

func (c *evalContextMock) GetReviewers() scm.Actors {
	args := c.Called()

	if actors, ok := args.Get(0).(scm.Actors); ok {
		return actors
	}

	return nil
}

func TestAssignReviewers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                      string
		step                      config.ActionStep
		mockGetReviewersResponse  scm.Actors
		mockGetCodeOwnersResponse scm.Actors
		wantUpdate                *scm.UpdateMergeRequestOptions
		wantErr                   error
	}{
		{
			name: "should not error on no source provided",
			step: config.ActionStep{
				"limit": 2,
			},
			mockGetReviewersResponse:  nil,
			mockGetCodeOwnersResponse: nil,
			wantUpdate:                &scm.UpdateMergeRequestOptions{},
			wantErr:                   nil,
		},
		{
			name: "should error on no limit provided",
			step: config.ActionStep{
				"source": "codeowners",
			},
			mockGetReviewersResponse:  nil,
			mockGetCodeOwnersResponse: nil,
			wantUpdate:                &scm.UpdateMergeRequestOptions{},
			wantErr:                   errors.New("Required 'step' key 'limit' is missing"),
		},
		{
			name: "should update reviewers with eligible codeowners",
			step: config.ActionStep{
				"source": "codeowners",
				"limit":  2,
				"mode":   "random",
			},
			mockGetCodeOwnersResponse: scm.Actors{
				{ID: "1", Username: "user1"},
				{ID: "2", Username: "user2"},
				{ID: "3", Username: "user3"},
			},
			wantUpdate: &scm.UpdateMergeRequestOptions{
				ReviewerIDs: scm.Ptr([]int{3, 2}),
			},
			wantErr: nil,
		},
		{
			name: "should update reviewers with eligible codeowners when limit is higher than eligible reviewers",
			step: config.ActionStep{
				"source": "codeowners",
				"limit":  6,
				"mode":   "random",
			},
			mockGetCodeOwnersResponse: scm.Actors{
				{ID: "1", Username: "user1"},
				{ID: "2", Username: "user2"},
				{ID: "3", Username: "user3"},
			},
			wantUpdate: &scm.UpdateMergeRequestOptions{
				ReviewerIDs: scm.Ptr([]int{1, 2, 3}),
			},
			wantErr: nil,
		},
		{
			name: "should not update reviewers if reviewers already assigned",
			step: config.ActionStep{
				"source": "codeowners",
				"limit":  2,
				"mode":   "random",
			},
			mockGetReviewersResponse: scm.Actors{
				{ID: "1", Username: "user1"},
				{ID: "2", Username: "user2"},
			},
			mockGetCodeOwnersResponse: scm.Actors{
				{ID: "3", Username: "user1"},
				{ID: "2", Username: "user2"},
				{ID: "1", Username: "user3"},
			},
			wantUpdate: &scm.UpdateMergeRequestOptions{},
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evalContext := new(evalContextMock)
			evalContext.On("GetReviewers").Return(tt.mockGetReviewersResponse)
			evalContext.On("GetCodeOwners").Return(tt.mockGetCodeOwnersResponse)

			client := &gitlab.Client{}
			update := &scm.UpdateMergeRequestOptions{}
			step := tt.step

			ctx := state.WithDryRun(context.Background(), false)
			ctx = state.WithRandomSeed(ctx, 1)

			err := client.AssignReviewers(ctx, evalContext, update, step)

			assert.Equal(t, tt.wantErr, err)

			if tt.wantUpdate.ReviewerIDs != nil {
				wantLimit := len(*tt.wantUpdate.ReviewerIDs)
				assert.Len(t, *update.ReviewerIDs, wantLimit)
				assert.EqualValues(t, tt.wantUpdate.ReviewerIDs, update.ReviewerIDs)
			}
		})
	}
}
