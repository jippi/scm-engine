package gitlab_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/integration/backstage"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jippi/scm-engine/testutils"
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

func (c *evalContextMock) GetAuthor() scm.Actor {
	args := c.Called()

	if actor, ok := args.Get(0).(scm.Actor); ok {
		return actor
	}

	return scm.Actor{}
}

func (c *evalContextMock) GetLabels() []string {
	args := c.Called()

	if labels, ok := args.Get(0).([]string); ok {
		return labels
	}

	return nil
}

func TestAssignReviewers_codeowners(t *testing.T) {
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
			name: "should error on no source provided",
			step: config.ActionStep{
				"limit": 2,
			},
			mockGetReviewersResponse:  nil,
			mockGetCodeOwnersResponse: nil,
			wantUpdate:                &scm.UpdateMergeRequestOptions{},
			wantErr:                   errors.New("Required 'step' key 'source' is missing"),
		},
		{
			name: "should not error on no limit provided",
			step: config.ActionStep{
				"source": "codeowners",
			},
			mockGetReviewersResponse:  nil,
			mockGetCodeOwnersResponse: nil,
			wantUpdate:                &scm.UpdateMergeRequestOptions{},
			wantErr:                   nil,
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
				ReviewerIDs: scm.Ptr([]int{1, 2}),
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

			ctx := state.WithDryRun(t.Context(), false)
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

func TestAssignReviewers_static(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		step                     config.ActionStep
		mockGetReviewersResponse scm.Actors
		wantUpdate               *scm.UpdateMergeRequestOptions
		wantErr                  error
	}{
		{
			name: "should assign static user ids",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100", "200", "300"},
				"limit":    2,
			},
			mockGetReviewersResponse: nil,
			wantUpdate: &scm.UpdateMergeRequestOptions{
				ReviewerIDs: scm.Ptr([]int{100, 200}),
			},
			wantErr: nil,
		},
		{
			name: "should assign all static user ids when limit exceeds count",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100", "200"},
				"limit":    5,
			},
			mockGetReviewersResponse: nil,
			wantUpdate: &scm.UpdateMergeRequestOptions{
				ReviewerIDs: scm.Ptr([]int{100, 200}),
			},
			wantErr: nil,
		},
		{
			name: "should error when user_ids is missing for static source",
			step: config.ActionStep{
				"source": "static",
			},
			mockGetReviewersResponse: nil,
			wantUpdate:               &scm.UpdateMergeRequestOptions{},
			wantErr:                  errors.New("Required 'step' key 'user_ids' is missing"),
		},
		{
			name: "should not assign if list members already satisfy the limit",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"50", "100"},
				"limit":    1,
			},
			mockGetReviewersResponse: scm.Actors{
				{ID: "50", Username: "existing"},
			},
			wantUpdate: &scm.UpdateMergeRequestOptions{},
			wantErr:    nil,
		},
		{
			name: "should assign single user with default limit",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100", "200", "300"},
			},
			mockGetReviewersResponse: nil,
			wantUpdate: &scm.UpdateMergeRequestOptions{
				ReviewerIDs: scm.Ptr([]int{100}),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evalContext := new(evalContextMock)
			evalContext.On("GetReviewers").Return(tt.mockGetReviewersResponse)

			client := &gitlab.Client{}
			update := &scm.UpdateMergeRequestOptions{}

			ctx := state.WithDryRun(t.Context(), false)
			ctx = state.WithRandomSeed(ctx, 1)

			err := client.AssignReviewers(ctx, evalContext, update, tt.step)

			assert.Equal(t, tt.wantErr, err)

			if tt.wantUpdate.ReviewerIDs != nil {
				wantLimit := len(*tt.wantUpdate.ReviewerIDs)
				require.NotNil(t, update.ReviewerIDs)
				assert.Len(t, *update.ReviewerIDs, wantLimit)
				assert.EqualValues(t, tt.wantUpdate.ReviewerIDs, update.ReviewerIDs)
			}
		})
	}
}

func TestAssignReviewers_staticMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		step                     config.ActionStep
		mockGetReviewersResponse scm.Actors
		wantReviewerIDs          *[]int
		wantErr                  error
	}{
		{
			name: "should assign listed user when no reviewers exist",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100"},
				"mode":     "static",
			},
			mockGetReviewersResponse: nil,
			wantReviewerIDs:          scm.Ptr([]int{100}),
			wantErr:                  nil,
		},
		{
			name: "should assign listed user and preserve existing reviewers",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100"},
				"mode":     "static",
			},
			mockGetReviewersResponse: scm.Actors{
				{ID: "50", Username: "existing"},
			},
			wantReviewerIDs: scm.Ptr([]int{50, 100}),
			wantErr:         nil,
		},
		{
			name: "should be a no-op when listed user is already a reviewer",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"50"},
				"mode":     "static",
			},
			mockGetReviewersResponse: scm.Actors{
				{ID: "50", Username: "existing"},
			},
			wantReviewerIDs: nil,
			wantErr:         nil,
		},
		{
			name: "should dedup listed users against existing reviewers",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"50", "100"},
				"mode":     "static",
			},
			mockGetReviewersResponse: scm.Actors{
				{ID: "50", Username: "existing"},
			},
			wantReviewerIDs: scm.Ptr([]int{50, 100}),
			wantErr:         nil,
		},
		{
			name: "should assign all listed users, ignoring limit",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100", "200", "300"},
				"mode":     "static",
				"limit":    2,
			},
			mockGetReviewersResponse: nil,
			wantReviewerIDs:          scm.Ptr([]int{100, 200, 300}),
			wantErr:                  nil,
		},
		{
			name: "should error when static mode is used with a non-static source",
			step: config.ActionStep{
				"source": "codeowners",
				"mode":   "static",
			},
			mockGetReviewersResponse: nil,
			wantReviewerIDs:          nil,
			wantErr:                  errors.New("step field 'mode: static' is only supported with 'source: static'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evalContext := new(evalContextMock)
			evalContext.On("GetReviewers").Return(tt.mockGetReviewersResponse)

			client := &gitlab.Client{}
			update := &scm.UpdateMergeRequestOptions{}

			ctx := state.WithDryRun(t.Context(), false)
			ctx = state.WithRandomSeed(ctx, 1)

			err := client.AssignReviewers(ctx, evalContext, update, tt.step)

			assert.Equal(t, tt.wantErr, err)
			assert.EqualValues(t, tt.wantReviewerIDs, update.ReviewerIDs)
		})
	}
}

func TestAssignReviewers_randomTopUp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		step                     config.ActionStep
		mockGetReviewersResponse scm.Actors
		wantReviewerIDs          *[]int
		wantErr                  error
	}{
		{
			name: "should top up from the list when an existing reviewer is not in the list",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100"},
				"mode":     "random",
			},
			mockGetReviewersResponse: scm.Actors{
				{ID: "50", Username: "not-in-list"},
			},
			wantReviewerIDs: scm.Ptr([]int{50, 100}),
			wantErr:         nil,
		},
		{
			name: "should count existing list members towards the limit and top up the remainder",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100", "200"},
				"mode":     "random",
				"limit":    2,
			},
			mockGetReviewersResponse: scm.Actors{
				{ID: "100", Username: "already-assigned"},
			},
			wantReviewerIDs: scm.Ptr([]int{100, 200}),
			wantErr:         nil,
		},
		{
			name: "should do nothing when list members already satisfy the limit",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100", "200"},
				"mode":     "random",
				"limit":    1,
			},
			mockGetReviewersResponse: scm.Actors{
				{ID: "100", Username: "already-assigned"},
			},
			wantReviewerIDs: nil,
			wantErr:         nil,
		},
		{
			name: "should do nothing when all list members are already assigned",
			step: config.ActionStep{
				"source":   "static",
				"user_ids": []string{"100"},
				"mode":     "random",
				"limit":    5,
			},
			mockGetReviewersResponse: scm.Actors{
				{ID: "100", Username: "already-assigned"},
			},
			wantReviewerIDs: nil,
			wantErr:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evalContext := new(evalContextMock)
			evalContext.On("GetReviewers").Return(tt.mockGetReviewersResponse)

			client := &gitlab.Client{}
			update := &scm.UpdateMergeRequestOptions{}

			ctx := state.WithDryRun(t.Context(), false)
			ctx = state.WithRandomSeed(ctx, 1)

			err := client.AssignReviewers(ctx, evalContext, update, tt.step)

			assert.Equal(t, tt.wantErr, err)
			assert.EqualValues(t, tt.wantReviewerIDs, update.ReviewerIDs)
		})
	}
}

func TestAssignReviewers_backstage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		step                     config.ActionStep
		mockGetReviewersResponse scm.Actors
		wantUpdate               *scm.UpdateMergeRequestOptions
		wantErr                  error
	}{
		{
			name: "should skip adding author to reviewers",
			step: config.ActionStep{
				"source": "backstage",
			},
			mockGetReviewersResponse: scm.Actors{},
			wantUpdate: &scm.UpdateMergeRequestOptions{
				ReviewerIDs: scm.Ptr([]int{2}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evalContext := new(evalContextMock)
			evalContext.On("GetAuthor").Return(scm.Actor{ID: "1", Username: "user1"})
			evalContext.On("GetReviewers").Return(tt.mockGetReviewersResponse)

			ctx := state.WithDryRun(t.Context(), false)
			ctx = state.WithBaseURL(ctx, "https://gitlab.example.com")
			ctx = state.WithToken(ctx, "token")
			ctx = state.WithProjectID(ctx, "group/test-system")
			ctx = state.WithRandomSeed(ctx, 1)

			r := testutils.GetRecorder(t)
			defer r.Stop()

			backstageClient, err := backstage.NewClient(ctx, "https://backstage.example.com", "", r.GetDefaultClient())
			if err != nil {
				t.Fatalf("failed to create backstage client: %v", err)
			}

			client, err := gitlab.NewClient(ctx, backstageClient)
			if err != nil {
				t.Fatalf("failed to create gitlab client: %v", err)
			}

			update := &scm.UpdateMergeRequestOptions{}

			err = client.AssignReviewers(ctx, evalContext, update, tt.step)

			assert.Equal(t, tt.wantErr, err)

			if tt.wantUpdate.ReviewerIDs != nil {
				wantLimit := len(*tt.wantUpdate.ReviewerIDs)
				require.NotNil(t, update.ReviewerIDs)
				assert.Len(t, *update.ReviewerIDs, wantLimit)
				assert.EqualValues(t, tt.wantUpdate.ReviewerIDs, update.ReviewerIDs)
			}
		})
	}
}
