package transport

import (
	"github.com/influxdata/telegraf"
)

// Transport interface for general transport functions
type Transport interface {
	SampleConfig() string
	telegraf.Initializer
}

// Receiver interface to get data that can be parsed in a later step
type Receiver interface {
	// Receive data from the endpoint(s)
	Receive() ([]byte, error)
	Transport
}
