package logger

import "io"

type creator func(cfg Config) (io.WriteCloser, error)

var loggerRegistry map[string]creator

func registerLogger(name string, loggerCreator creator) {
	if loggerRegistry == nil {
		loggerRegistry = make(map[string]creator)
	}
	loggerRegistry[name] = loggerCreator
}
