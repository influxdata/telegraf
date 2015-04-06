package plugins

type Accumulator interface {
	Add(name string, value interface{}, tags map[string]string)
}

type Plugin interface {
	Gather(Accumulator) error
}

type Creator func() Plugin

var Plugins = map[string]Creator{}

func Add(name string, creator Creator) {
	Plugins[name] = creator
}
