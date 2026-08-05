package backstage_test

import (
	"errors"
	"testing"

	go_backstage "github.com/datolabs-io/go-backstage/v3"
	"github.com/jippi/scm-engine/pkg/integration/backstage"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"github.com/jippi/scm-engine/testutils"
)

func TestClient_GetEntityOwner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		filters []string
		want    *backstage.EntityReference
		wantErr error
	}{
		{
			name:    "not found",
			filters: []string{"kind=system,metadata.name=example"},
			wantErr: errors.New("No system found in Backstage catalog"),
		},
		{
			name:    "found",
			filters: []string{"kind=system,metadata.name=test-system"},
			want: &backstage.EntityReference{
				Kind:      "group",
				Namespace: "default",
				Name:      "test-group",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := testutils.GetRecorder(t)
			defer r.Stop()

			client, err := backstage.NewClient(t.Context(), "https://backstage.example.com", "", r.GetDefaultClient())
			require.NoError(t, err)

			entityRef, err := client.GetEntityOwner(t.Context(), tt.filters...)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, tt.wantErr, err.Error())

				return
			}

			require.NoError(t, err)
			assert.DeepEqual(t, tt.want, entityRef)
		})
	}
}

func TestClient_GetUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		arg     *backstage.EntityReference
		want    *go_backstage.Entity
		wantErr error
	}{
		{
			name: "found",
			arg: &backstage.EntityReference{
				Name:      "test-user",
				Namespace: "default",
			},
			want: &go_backstage.Entity{
				Metadata: go_backstage.EntityMeta{
					UID:       "00000000-0000-0000-0000-000000000000",
					Etag:      "00000000000000000",
					Name:      "test-user",
					Namespace: "default",
					Labels:    map[string]string{},
					Annotations: map[string]string{
						"gitlab.com/user_id": "1",
					},
				},
			},
		},
		{
			name: "not found",
			arg: &backstage.EntityReference{
				Name:      "missing-user",
				Namespace: "default",
			},
			wantErr: errors.New("Failed to get user: 404 Not Found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := testutils.GetRecorder(t)
			defer r.Stop()

			client, err := backstage.NewClient(t.Context(), "https://backstage.example.com", "", r.GetDefaultClient())
			require.NoError(t, err)

			user, err := client.GetUser(t.Context(), tt.arg)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, tt.wantErr, err.Error())

				return
			}

			require.NoError(t, err)
			assert.DeepEqual(t, *tt.want, user.Entity)
		})
	}
}

func TestClient_ListGroupMembers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		arg     *backstage.EntityReference
		want    []go_backstage.Entity
		wantErr error
	}{
		{
			name: "found",
			arg: &backstage.EntityReference{
				Kind:      "group",
				Name:      "test-group",
				Namespace: "default",
			},
			want: []go_backstage.Entity{
				{
					ApiVersion: "backstage.io/v1alpha1",
					Kind:       "User",
					Spec: map[string]any{
						"profile": map[string]any{
							"displayName": "Test User",
							"email":       "test-user@example.com",
						},
					},
					Metadata: go_backstage.EntityMeta{
						UID:       "00000000-0000-0000-0000-000000000000",
						Etag:      "0",
						Name:      "test-user",
						Namespace: "default",
						Labels:    map[string]string{},
						Annotations: map[string]string{
							"gitlab.com/user_id": "1",
						},
					},
				},
			},
		},
		{
			name: "not found",
			arg: &backstage.EntityReference{
				Name:      "missing-group",
				Namespace: "default",
			},
			want: []go_backstage.Entity{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := testutils.GetRecorder(t)
			defer r.Stop()

			client, err := backstage.NewClient(t.Context(), "https://backstage.example.com", "", r.GetDefaultClient())
			require.NoError(t, err)

			members, err := client.ListGroupMembers(t.Context(), tt.arg)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, tt.wantErr, err.Error())

				return
			}

			require.NoError(t, err)
			assert.DeepEqual(t, tt.want, members)
		})
	}
}

func TestClient_GetOwnersForGitLabProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		arg     string
		want    []scm.Actor
		wantErr error
	}{
		{
			name: "found",
			arg:  "test-system",
			want: []scm.Actor{
				{
					ID: "1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := testutils.GetRecorder(t)
			defer r.Stop()

			client, err := backstage.NewClient(t.Context(), "https://backstage.example.com", "", r.GetDefaultClient())
			require.NoError(t, err)

			owners, err := client.GetOwnersForGitLabProject(t.Context(), tt.arg)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, tt.wantErr, err.Error())

				return
			}

			require.NoError(t, err)
			assert.DeepEqual(t, tt.want, owners)
		})
	}
}
