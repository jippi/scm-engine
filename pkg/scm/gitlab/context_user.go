package gitlab

import "github.com/jippi/scm-engine/pkg/scm"

func (u ContextUser) ToActor() scm.Actor {
	return scm.Actor{
		Username: u.Username,
		IsBot:    u.Bot,
		Email:    u.PublicEmail,
	}
}
