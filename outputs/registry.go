package outputs

import (
	"github.com/influxdb/influxdb/client/v2"
)

type Output interface {
	Connect() error
	Close() error
	Description() string
	SampleConfig() string
	Write(points []*client.Point) error
}

type Creator func() Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
