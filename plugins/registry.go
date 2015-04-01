package plugins

import "github.com/vektra/cypress"

type Plugin interface {
	Read() ([]*cypress.Message, error)
}

type Creator func() Plugin

var Plugins = map[string]Creator{}

func Add(name string, creator Creator) {
	Plugins[name] = creator
}
