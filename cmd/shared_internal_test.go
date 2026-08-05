//nolint:testpackage // syncLabels, runActions and updateMergeRequest are unexported, and they decide which API calls scm-engine makes
package cmd

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/require"
)

// fakeLabelClient records every write so a test can assert on exactly which
// labels were created or updated.
type fakeLabelClient struct {
	existing []*scm.Label
	listErr  error

	created []string
	updated []string

	createErr    error
	createStatus int
	updateErr    error
}

func (c *fakeLabelClient) List(context.Context) ([]*scm.Label, error) {
	return c.existing, c.listErr
}

func (c *fakeLabelClient) Create(_ context.Context, opt *scm.CreateLabelOptions) (*scm.Label, *scm.Response, error) {
	c.created = append(c.created, *opt.Name)

	if c.createErr != nil {
		status := c.createStatus
		if status == 0 {
			status = http.StatusInternalServerError
		}

		return nil, &scm.Response{Response: &http.Response{StatusCode: status}}, c.createErr
	}

	return nil, &scm.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
}

func (c *fakeLabelClient) Update(_ context.Context, opt *scm.UpdateLabelOptions) (*scm.Label, *scm.Response, error) {
	c.updated = append(c.updated, *opt.Name)

	return nil, nil, c.updateErr
}

// fakeMergeRequestClient records whether an update was actually sent.
type fakeMergeRequestClient struct {
	updates   []*scm.UpdateMergeRequestOptions
	updateErr error
}

func (c *fakeMergeRequestClient) Update(_ context.Context, opt *scm.UpdateMergeRequestOptions) (*scm.Response, error) {
	c.updates = append(c.updates, opt)

	return nil, c.updateErr
}

func (c *fakeMergeRequestClient) List(context.Context, *scm.ListMergeRequestsOptions) ([]scm.ListMergeRequest, error) {
	return nil, errNotImplemented
}

func (c *fakeMergeRequestClient) GetRemoteConfig(context.Context, string, string) (io.Reader, error) {
	return nil, errNotImplemented
}

type fakeClient struct {
	labels        *fakeLabelClient
	mergeRequests *fakeMergeRequestClient

	appliedSteps []scm.ActionStep
	applyErr     error
}

func newFakeClient() *fakeClient {
	return &fakeClient{
		labels:        &fakeLabelClient{},
		mergeRequests: &fakeMergeRequestClient{},
	}
}

func (c *fakeClient) Labels() scm.LabelClient               { return c.labels }
func (c *fakeClient) MergeRequests() scm.MergeRequestClient { return c.mergeRequests }

func (c *fakeClient) ApplyStep(_ context.Context, _ scm.EvalContext, _ *scm.UpdateMergeRequestOptions, step scm.ActionStep) error {
	c.appliedSteps = append(c.appliedSteps, step)

	return c.applyErr
}

// errNotImplemented is returned by the parts of the client interface these
// tests never reach, so a future test that does reach them fails loudly.
var errNotImplemented = errors.New("not implemented by the test fake")

func (c *fakeClient) EvalContext(context.Context) (scm.EvalContext, error) {
	return nil, errNotImplemented
}

func (c *fakeClient) FindMergeRequestsForPeriodicEvaluation(context.Context, scm.MergeRequestListFilters) ([]scm.PeriodicEvaluationMergeRequest, error) {
	return nil, errNotImplemented
}

func (c *fakeClient) GetProjectFiles(context.Context, string, *string, []string) (map[string]string, error) {
	return nil, errNotImplemented
}

func (c *fakeClient) Start(context.Context) error             { return nil }
func (c *fakeClient) Stop(context.Context, error, bool) error { return nil }

// evalContextStub tracks action groups the same way the real contexts do.
type evalContextStub struct {
	scm.EvalContext

	groups map[string]bool
}

func newEvalContextStub() *evalContextStub {
	return &evalContextStub{groups: map[string]bool{}}
}

func (c *evalContextStub) TrackActionGroupExecution(name string) {
	if len(name) == 0 {
		return
	}

	c.groups[name] = true
}

func (c *evalContextStub) HasExecutedActionGroup(name string) bool {
	if len(name) == 0 {
		return false
	}

	return c.groups[name]
}

// An empty update must never be sent: GitLab returns an error for a PUT with no
// changes, and it burns an API call on every evaluation that changes nothing.
func TestUpdateMergeRequest_skipsEmptyUpdates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		update *scm.UpdateMergeRequestOptions
	}{
		{name: "nil update", update: nil},
		{name: "empty update", update: &scm.UpdateMergeRequestOptions{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newFakeClient()
			ctx := state.WithDryRun(t.Context(), false)

			require.NoError(t, updateMergeRequest(ctx, client, tt.update))
			require.Empty(t, client.mergeRequests.updates, "no request should have been sent")
		})
	}
}

func TestUpdateMergeRequest_sendsRealChanges(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	ctx := state.WithDryRun(t.Context(), false)
	update := &scm.UpdateMergeRequestOptions{StateEvent: scm.Ptr("close")}

	require.NoError(t, updateMergeRequest(ctx, client, update))
	require.Len(t, client.mergeRequests.updates, 1)
	require.Equal(t, update, client.mergeRequests.updates[0])
}

func TestUpdateMergeRequest_dryRunSendsNothing(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	ctx := state.WithDryRun(t.Context(), true)

	require.NoError(t, updateMergeRequest(ctx, client, &scm.UpdateMergeRequestOptions{StateEvent: scm.Ptr("close")}))
	require.Empty(t, client.mergeRequests.updates)
}

func TestUpdateMergeRequest_propagatesError(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.mergeRequests.updateErr = errors.New("boom")

	ctx := state.WithDryRun(t.Context(), false)

	require.ErrorContains(t,
		updateMergeRequest(ctx, client, &scm.UpdateMergeRequestOptions{StateEvent: scm.Ptr("close")}),
		"boom")
}

func TestRunActions_noActions(t *testing.T) {
	t.Parallel()

	client := newFakeClient()

	require.NoError(t, runActions(t.Context(), newEvalContextStub(), client, &scm.UpdateMergeRequestOptions{}, nil))
	require.Empty(t, client.appliedSteps)
}

func TestRunActions_appliesEveryStep(t *testing.T) {
	t.Parallel()

	client := newFakeClient()

	actions := config.Actions{
		{Name: "first", Then: []config.ActionStep{{"action": "close"}, {"action": "comment"}}},
		{Name: "second", Then: []config.ActionStep{{"action": "approve"}}},
	}

	require.NoError(t, runActions(t.Context(), newEvalContextStub(), client, &scm.UpdateMergeRequestOptions{}, actions))
	require.Len(t, client.appliedSteps, 3)
}

// Only the first action in a group may run per evaluation.
func TestRunActions_skipsRepeatedGroup(t *testing.T) {
	t.Parallel()

	client := newFakeClient()

	actions := config.Actions{
		{Name: "first", Group: "stale", Then: []config.ActionStep{{"action": "comment"}}},
		{Name: "second", Group: "stale", Then: []config.ActionStep{{"action": "close"}}},
		{Name: "third", Group: "other", Then: []config.ActionStep{{"action": "approve"}}},
	}

	require.NoError(t, runActions(t.Context(), newEvalContextStub(), client, &scm.UpdateMergeRequestOptions{}, actions))

	require.Len(t, client.appliedSteps, 2)

	// The second "stale" action is skipped, so the surviving steps are the
	// first action's comment and the differently grouped approve.
	actionNames := make([]string, 0, len(client.appliedSteps))

	for _, step := range client.appliedSteps {
		name, err := step.RequiredString("action")
		require.NoError(t, err)

		actionNames = append(actionNames, name)
	}

	require.Equal(t, []string{"comment", "approve"}, actionNames)
}

// Actions without a group are independent and must all run.
func TestRunActions_ungroupedActionsAllRun(t *testing.T) {
	t.Parallel()

	client := newFakeClient()

	actions := config.Actions{
		{Name: "first", Then: []config.ActionStep{{"action": "comment"}}},
		{Name: "second", Then: []config.ActionStep{{"action": "close"}}},
		{Name: "third", Then: []config.ActionStep{{"action": "approve"}}},
	}

	require.NoError(t, runActions(t.Context(), newEvalContextStub(), client, &scm.UpdateMergeRequestOptions{}, actions))
	require.Len(t, client.appliedSteps, 3)
}

func TestRunActions_stopsOnError(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.applyErr = errors.New("step failed")

	actions := config.Actions{
		{Name: "first", Then: []config.ActionStep{{"action": "close"}}},
		{Name: "second", Then: []config.ActionStep{{"action": "approve"}}},
	}

	require.ErrorContains(t,
		runActions(t.Context(), newEvalContextStub(), client, &scm.UpdateMergeRequestOptions{}, actions),
		"step failed")

	require.Len(t, client.appliedSteps, 1, "evaluation must stop at the failing step")
}

func TestSyncLabels_createsMissingLabels(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.labels.existing = []*scm.Label{{Name: "existing", Color: "#FF0000"}}

	ctx := state.WithDryRun(state.WithProvider(t.Context(), "gitlab"), false)

	required := []scm.EvaluationResult{
		{Name: "existing", Color: "#FF0000"},
		{Name: "brand-new", Color: "#00FF00"},
	}

	require.NoError(t, syncLabels(ctx, client, required))
	require.Equal(t, []string{"brand-new"}, client.labels.created)
}

// A label that already matches must not be rewritten on every evaluation.
func TestSyncLabels_skipsUnchangedLabels(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.labels.existing = []*scm.Label{{Name: "bug", Color: "#FF0000", Description: "a bug"}}

	ctx := state.WithDryRun(state.WithProvider(t.Context(), "gitlab"), false)

	require.NoError(t, syncLabels(ctx, client, []scm.EvaluationResult{
		{Name: "bug", Color: "#FF0000", Description: "a bug"},
	}))

	require.Empty(t, client.labels.created)
	require.Empty(t, client.labels.updated, "an unchanged label must not be updated")
}

func TestSyncLabels_updatesChangedLabels(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.labels.existing = []*scm.Label{{Name: "bug", Color: "#FF0000", Description: "old"}}

	ctx := state.WithDryRun(state.WithProvider(t.Context(), "gitlab"), false)

	require.NoError(t, syncLabels(ctx, client, []scm.EvaluationResult{
		{Name: "bug", Color: "#FF0000", Description: "new"},
	}))

	require.Empty(t, client.labels.created)
	require.Equal(t, []string{"bug"}, client.labels.updated)
}

func TestSyncLabels_dryRunWritesNothing(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.labels.existing = []*scm.Label{{Name: "bug", Description: "old"}}

	ctx := state.WithDryRun(state.WithProvider(t.Context(), "gitlab"), true)

	require.NoError(t, syncLabels(ctx, client, []scm.EvaluationResult{
		{Name: "bug", Description: "new"},
		{Name: "brand-new"},
	}))

	require.Empty(t, client.labels.created)
	require.Empty(t, client.labels.updated)
}

// Two evaluations racing to create the same label is expected; the loser gets a
// 409 and must carry on rather than failing the whole run.
func TestSyncLabels_toleratesConflictOnCreate(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.labels.createErr = errors.New("already exists")
	client.labels.createStatus = http.StatusConflict

	ctx := state.WithDryRun(state.WithProvider(t.Context(), "gitlab"), false)

	require.NoError(t, syncLabels(ctx, client, []scm.EvaluationResult{{Name: "one"}, {Name: "two"}}))
	require.Equal(t, []string{"one", "two"}, client.labels.created, "a conflict must not stop the remaining labels")
}

func TestSyncLabels_failsOnOtherCreateErrors(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.labels.createErr = errors.New("server exploded")
	client.labels.createStatus = http.StatusInternalServerError

	ctx := state.WithDryRun(state.WithProvider(t.Context(), "gitlab"), false)

	require.ErrorContains(t, syncLabels(ctx, client, []scm.EvaluationResult{{Name: "one"}}), "server exploded")
}

func TestSyncLabels_propagatesListError(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.labels.listErr = errors.New("cannot list")

	ctx := state.WithDryRun(state.WithProvider(t.Context(), "gitlab"), false)

	require.ErrorContains(t, syncLabels(ctx, client, []scm.EvaluationResult{{Name: "one"}}), "cannot list")
}

func TestSyncLabels_propagatesUpdateError(t *testing.T) {
	t.Parallel()

	client := newFakeClient()
	client.labels.existing = []*scm.Label{{Name: "bug", Description: "old"}}
	client.labels.updateErr = errors.New("cannot update")

	ctx := state.WithDryRun(state.WithProvider(t.Context(), "gitlab"), false)

	require.ErrorContains(t, syncLabels(ctx, client, []scm.EvaluationResult{{Name: "bug", Description: "new"}}), "cannot update")
}
