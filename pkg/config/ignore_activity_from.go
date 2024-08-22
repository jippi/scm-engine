package config

import "github.com/jippi/scm-engine/pkg/scm"

type IgnoreActivityFrom struct {
	IsBot     bool     `yaml:"bots"`
	Usernames []string `yaml:"usernames"`
	Emails    []string `yaml:"emails"`
}

func (i IgnoreActivityFrom) Matches(actor scm.Actor) bool {
	// If actor is bot and we ignore bot activity
	if actor.IsBot && i.IsBot {
		return true
	}

	// Check if the actor username is in the ignore list
	for _, username := range i.Usernames {
		if username == actor.Username {
			return true
		}
	}

	// If the actor don't have an email, we did not find a match, since
	// our last check is on emails
	if actor.Email == nil {
		return false
	}

	// Check if the actor email matches any of the ignored ones
	for _, email := range i.Emails {
		if email == *actor.Email {
			return true
		}
	}

	return false
}
