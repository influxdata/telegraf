package telegraf

import "github.com/influxdata/telegraf/pubsub"

type Service interface {
	// Connect to the Output
	Connect() error
	// Close any connections to the Output
	Close() error
	// Description returns a one-sentence description on the Output
	Description() string
	// SampleConfig returns the default configuration of the Output
	SampleConfig() string
	// Write takes in group of points to be written to the Output
	Run(msgbus *pubsub.PubSub) error
}
