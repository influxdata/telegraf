package microsoft_fabric

import "github.com/influxdata/telegraf"

type FabricOutput interface {
	Init() error
	Connect() error
	Write(metrics []telegraf.Metric) error
	Close() error
}
