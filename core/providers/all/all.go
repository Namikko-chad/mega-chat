// Package all builds the registry of every provider the core ships with.
package all

import (
	"megachat/core/provider"
	"megachat/core/providers/fake"
	"megachat/core/providers/google"
	"megachat/core/providers/mailru"
	"megachat/core/providers/slack"
	"megachat/core/providers/telegram"
	"megachat/core/providers/vkteams"
)

// Registry returns the providers; demo adds the in-memory "fake" one for
// working on the app without real accounts.
func Registry(demo bool) *provider.Registry {
	r := provider.NewRegistry(telegram.New(), slack.New(), vkteams.New(), google.New(), mailru.New())
	if demo {
		r.Register(fake.New())
	}
	return r
}
