package config

import "github.com/jippi/scm-engine/pkg/scm"

type IgnoreActivityFrom struct {
	// (Optional) Should bot users be ignored when considering activity? Default: false
	//
	// See: https://jippi.github.io/scm-engine/configuration/#ignore_activity_from.bots
	IsBot bool `json:"bots,omitempty" yaml:"bots" jsonschema:"default=false"`

	// (Optional) A list of usernames that should be ignored when considering user activity. Default: []
	//
	// See: https://jippi.github.io/scm-engine/configuration/#ignore_activity_from.usernames
	Usernames []string `json:"usernames,omitempty" yaml:"usernames"`

	// (Optional) A list of emails that should be ignored when considering user activity. Default: []
	// NOTE: If a user do not have a public email configured on their profile, that users activity will never match this rule.
	//
	// See: https://jippi.github.io/scm-engine/configuration/#ignore_activity_from.emails
	Emails []string `json:"emails,omitempty" yaml:"emails"`
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
