// Package all builds the registry of every provider the core ships with.
package all

import (
	"github.com/SelfHold/parley/core/provider"
	"github.com/SelfHold/parley/core/providers/fake"
	"github.com/SelfHold/parley/core/providers/google"
	"github.com/SelfHold/parley/core/providers/mailru"
	"github.com/SelfHold/parley/core/providers/slack"
	"github.com/SelfHold/parley/core/providers/telegram"
	"github.com/SelfHold/parley/core/providers/vk"
	"github.com/SelfHold/parley/core/providers/vkteams"
)

// Registry returns the providers; demo adds the in-memory "fake" one for
// working on the app without real accounts.
func Registry(demo bool) *provider.Registry {
	r := provider.NewRegistry(telegram.New(), slack.New(), vk.New(), vkteams.New(), google.New(), mailru.New())
	if demo {
		f := fake.New()
		f.Special = true
		r.Register(f)
	}
	return r
}
