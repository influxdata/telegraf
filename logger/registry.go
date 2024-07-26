package logger

type creator func(cfg *Config) (sink, error)

var registry = make(map[string]creator)

func add(name string, creator creator) {
	registry[name] = creator
}
