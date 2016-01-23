package outputs

import (
	"github.com/influxdata/influxdb/client/v2"
)

type Output interface {
	// Connect to the Output
	Connect() error
	// Close any connections to the Output
	Close() error
	// Description returns a one-sentence description on the Output
	Description() string
	// SampleConfig returns the default configuration of the Output
	SampleConfig() string
	// Write takes in group of points to be written to the Output
	Write(points []*client.Point) error
}

type ServiceOutput interface {
	// Connect to the Output
	Connect() error
	// Close any connections to the Output
	Close() error
	// Description returns a one-sentence description on the Output
	Description() string
	// SampleConfig returns the default configuration of the Output
	SampleConfig() string
	// Write takes in group of points to be written to the Output
	Write(points []*client.Point) error
	// Start the "service" that will provide an Output
	Start() error
	// Stop the "service" that will provide an Output
	Stop()
}

type Creator func() Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
