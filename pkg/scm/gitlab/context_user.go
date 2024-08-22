package gitlab

import "github.com/jippi/scm-engine/pkg/config"

func (u ContextUser) ToActorMatcher() config.ActorMatcher {
	return config.ActorMatcher{
		Username: u.Username,
		IsBot:    u.Bot,
		Email:    u.PublicEmail,
	}
}
