package telegraf

// LogLevel denotes the level for logging
type LogLevel int

const (
	// None means nothing is logged
	None LogLevel = iota
	// Error will log error messages
	Error
	// Warn will log error messages and warnings
	Warn
	// Info will log error messages, warnings and information messages
	Info
	// Debug will log all of the above and debugging messages issued by plugins
	Debug
)

func LogLevelFromString(name string) LogLevel {
	switch name {
	case "ERROR", "error":
		return Error
	case "WARN", "warn":
		return Warn
	case "INFO", "info":
		return Info
	case "DEBUG", "debug":
		return Debug
	}
	return None
}

func (e LogLevel) String() string {
	switch e {
	case Error:
		return "ERROR"
	case Warn:
		return "WARN"
	case Info:
		return "INFO"
	case Debug:
		return "DEBUG"
	}
	return "NONE"
}

func (e LogLevel) Indicator() string {
	switch e {
	case Error:
		return "E!"
	case Warn:
		return "W!"
	case Info:
		return "I!"
	case Debug:
		return "D!"
	}
	return "U!"
}

func (e LogLevel) Includes(level LogLevel) bool {
	return e >= level
}

// Logger defines an plugin-related interface for logging.
type Logger interface {
	// Level returns the configured log-level of the logger
	Level() LogLevel

	// Errorf logs an error message, patterned after log.Printf.
	Errorf(format string, args ...interface{})
	// Error logs an error message, patterned after log.Print.
	Error(args ...interface{})
	// Debugf logs a debug message, patterned after log.Printf.
	Debugf(format string, args ...interface{})
	// Debug logs a debug message, patterned after log.Print.
	Debug(args ...interface{})
	// Warnf logs a warning message, patterned after log.Printf.
	Warnf(format string, args ...interface{})
	// Warn logs a warning message, patterned after log.Print.
	Warn(args ...interface{})
	// Infof logs an information message, patterned after log.Printf.
	Infof(format string, args ...interface{})
	// Info logs an information message, patterned after log.Print.
	Info(args ...interface{})
}
