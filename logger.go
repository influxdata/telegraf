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
	// Trace will log all of the above and trace messages issued by plugins
	Trace
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
	case "TRACE", "trace":
		return Trace
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
	case Trace:
		return "TRACE"
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
	case Trace:
		return "T!"
	}
	return "U!"
}

func (e LogLevel) Includes(level LogLevel) bool {
	return e >= level
}

// Logger defines an plugin-related interface for logging.
type Logger interface { //nolint:interfacebloat // All functions are required
	// Level returns the configured log-level of the logger
	Level() LogLevel

	// AddAttribute allows to add a key-value attribute to the logging output
	AddAttribute(key string, value any)

	// Errorf logs an error message, patterned after log.Printf.
	Errorf(format string, args ...any)
	// Error logs an error message, patterned after log.Print.
	Error(args ...any)
	// Warnf logs a warning message, patterned after log.Printf.
	Warnf(format string, args ...any)
	// Warn logs a warning message, patterned after log.Print.
	Warn(args ...any)
	// Infof logs an information message, patterned after log.Printf.
	Infof(format string, args ...any)
	// Info logs an information message, patterned after log.Print.
	Info(args ...any)
	// Debugf logs a debug message, patterned after log.Printf.
	Debugf(format string, args ...any)
	// Debug logs a debug message, patterned after log.Print.
	Debug(args ...any)
	// Tracef logs a trace message, patterned after log.Printf.
	Tracef(format string, args ...any)
	// Trace logs a trace message, patterned after log.Print.
	Trace(args ...any)
}
