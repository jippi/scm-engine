package config_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/stretchr/testify/require"
)

func TestIgnoreActivityFrom_Matches(t *testing.T) {
	t.Parallel()

	defaultActor := scm.Actor{
		Username: "jippi",
		IsBot:    false,
		Email:    scm.Ptr("jippi@scm-engine.example.com"),
	}

	botActor := scm.Actor{
		Username: "scm-engine",
		IsBot:    true,
	}

	tests := []struct {
		name   string
		actor  scm.Actor
		fields config.IgnoreActivityFrom
		want   bool
	}{
		{
			name: "empty",
			want: false,
		},
		{
			name:  "default actor",
			actor: defaultActor,
		},
		{
			name:   "username: matching username",
			actor:  defaultActor,
			fields: config.IgnoreActivityFrom{Usernames: []string{"jippi"}},
			want:   true,
		},
		{
			name:   "username: partial matching",
			actor:  defaultActor,
			fields: config.IgnoreActivityFrom{Usernames: []string{"jippignu"}},
			want:   false,
		},
		{
			name:   "bot:, actor not a bot",
			actor:  defaultActor,
			fields: config.IgnoreActivityFrom{IsBot: true},
			want:   false,
		},
		{
			name:   "bot: ignore bot, actor is a bot",
			actor:  botActor,
			fields: config.IgnoreActivityFrom{IsBot: true},
			want:   true,
		},
		{
			name:   "ignore email, actor without email, should not match",
			actor:  scm.Actor{Username: "jippi"},
			fields: config.IgnoreActivityFrom{Emails: []string{"demo@example.com"}},
			want:   false,
		},
		{
			name:   "email: actor with different email, should not match",
			actor:  defaultActor,
			fields: config.IgnoreActivityFrom{Emails: []string{"demo@example.com"}},
			want:   false,
		},
		{
			name:   "email: actor with matching email",
			actor:  defaultActor,
			fields: config.IgnoreActivityFrom{Emails: []string{"jippi@scm-engine.example.com"}},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tt.fields.Matches(tt.actor))
		})
	}
}
