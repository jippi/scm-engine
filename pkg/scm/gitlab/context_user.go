package gitlab

import "github.com/jippi/scm-engine/pkg/config"

func (u ContextUser) ToActor() config.Actor {
	return config.Actor{
		Username: u.Username,
		IsBot:    u.Bot,
		Email:    u.PublicEmail,
	}
}
