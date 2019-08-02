package telegraf

import "log"

// Initializer is an interface that all plugin types: Inputs, Outputs,
// Processors, and Aggregators can optionally implement to initialize the
// plugin.
type Initializer interface {
	// Init performs one time setup of the plugin and returns an error if the
	// configuration is invalid.
	Init(PluginConfig) error
}

type Input interface {
	// SampleConfig returns the default configuration of the Input
	SampleConfig() string

	// Description returns a one-sentence description on the Input
	Description() string

	// Gather takes in an accumulator and adds the metrics that the Input
	// gathers. This is called every "interval"
	Gather(Accumulator) error
}

type ServiceInput interface {
	Input

	// Start the ServiceInput.  The Accumulator may be retained and used until
	// Stop returns.
	Start(Accumulator) error

	// Stop stops the services and closes any necessary channels and connections
	Stop()
}

// PluginConfig contains individualized plugin configuration.
type PluginConfig struct {
	Logger Logger
}

// Logger defines a logging structure for plugins.
type Logger struct {
	Name string // Name is the plugin name, will be printed in the `[]`.
}

// Errorf logs an error message, patterned after log.Printf.
func (l Logger) Errorf(format string, args ...interface{}) {
	// todo: keep tally of errors from plugins
	log.Printf("E! ["+l.Name+"] "+format, args...)
}

// Error logs an error message, patterned after log.Print.
func (l Logger) Error(args ...interface{}) {
	log.Print(append([]interface{}{"E! [" + l.Name + "] "}, args...)...)
}

// Debugf logs a debug message, patterned after log.Printf.
func (l Logger) Debugf(format string, args ...interface{}) {
	log.Printf("D! ["+l.Name+"] "+format, args...)
}

// Debug logs a debug message, patterned after log.Print.
func (l Logger) Debug(args ...interface{}) {
	log.Print(append([]interface{}{"D! [" + l.Name + "] "}, args...)...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (l Logger) Warnf(format string, args ...interface{}) {
	log.Printf("W! ["+l.Name+"] "+format, args...)
}

// Warn logs a warning message, patterned after log.Print.
func (l Logger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! [" + l.Name + "] "}, args...)...)
}

// Infof logs an information message, patterned after log.Printf.
func (l Logger) Infof(format string, args ...interface{}) {
	log.Printf("I! ["+l.Name+"] "+format, args...)
}

// Info logs an information message, patterned after log.Print.
func (l Logger) Info(args ...interface{}) {
	log.Print(append([]interface{}{"I! [" + l.Name + "] "}, args...)...)
}
