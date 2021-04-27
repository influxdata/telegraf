package telegraf

var Debug bool

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
	// SampleConfig returns the default configuration of the Processor
	SampleConfig() string

	// Description returns a one-sentence description on the Processor
	Description() string
}

// StatefulPluginWithID allows a plugin to overwrite the state
// identifier of the plugin instance. By default the state-persister
// will generate an ID for the plugin depending on the plugin's
// configuration. However, it might be favorable to set the ID in
// another way, e.g. by allowing the user to set an ID in the config.
type StatefulPluginWithID interface {
	// GetStateID returns the ID of the plugin instance
	// Note: This function has to be callable directly after the
	// plugin's Init() function if there is any!
	GetPluginStateID() string

	StatefulPlugin
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

// StatePersister defines the plugin facing interface
// for persisting states
type StatePersister interface {
	// UpdateState can be called by a plugin to actively announce
	// a changed state to the Persister.
	// Note: The persister may or may not immediately persist
	// the state to disk.
	UpdateState(state interface{}) error
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
