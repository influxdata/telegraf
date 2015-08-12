package outputs

import (
	"github.com/influxdb/influxdb/client"
)

type Output interface {
	Connect() error
	Close() error
	Write(client.BatchPoints) error
}

type Creator func() Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
