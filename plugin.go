package telegraf

var Debug bool

// Escalation level for the plugin or option
type Escalation int

func (e Escalation) String() string {
	switch e {
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	}
	return "NONE"
}

const (
	// None means no deprecation
	None Escalation = iota
	// Warn means deprecated but still within the grace period
	Warn
	// Error means deprecated and beyond grace period
	Error
)

// DeprecationInfo contains information for marking a plugin deprecated.
type DeprecationInfo struct {
	// Since specifies the version since when the plugin is deprecated
	Since string
	// RemovalIn optionally specifies the version when the plugin is scheduled for removal
	RemovalIn string
	// Notice for the user on suggested replacements etc.
	Notice string
}

// Initializer is an interface that all plugin types: Inputs, Outputs,
// Processors, and Aggregators can optionally implement to initialize the
// plugin.
type Initializer interface {
	// Init performs one time setup of the plugin and returns an error if the
	// configuration is invalid.
	Init() error
}

// PluginDescriber contains the functions all plugins must implement to describe
// themselves to Telegraf. Note that all plugins may define a logger that is
// not part of the interface, but will receive an injected logger if it's set.
// eg: Log telegraf.Logger `toml:"-"`
type PluginDescriber interface {
	// SampleConfig returns the default configuration of the Plugin
	SampleConfig() string
}

// PluginWithID allows a plugin to overwrite its identifier of the plugin
// instance by a user specified value. By default the ID is generated
// using the plugin's configuration.
type PluginWithID interface {
	// ID returns the ID of the plugin instance. This function has to be
	// callable directly after the plugin's Init() function if there is any!
	ID() string
}

// StatefulPlugin contains the functions that plugins must implement to
// persist an internal state across Telegraf runs.
// Note that plugins may define a persister that is not part of the
// interface, but can be used to trigger state updates by the plugin if
// it exists in the plugin struct,
// eg: Persister telegraf.StatePersister `toml:"-"`
type StatefulPlugin interface {
	// GetState returns the current state of the plugin to persist
	// The returned state can be of any time as long as it can be
	// serialized to JSON. The best choice is a structure defined in
	// your plugin.
	// Note: This function has to be callable directly after the
	// plugin's Init() function if there is any!
	GetState() interface{}

	// SetState is called by the Persister once after loading and
	// initialization (after Init() function).
	SetState(state interface{}) error
}

// Logger defines an plugin-related interface for logging.
type Logger interface {
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
