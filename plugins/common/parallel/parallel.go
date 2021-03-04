package parallel

import "github.com/influxdata/telegraf"

type Parallel interface {
	Enqueue(telegraf.Metric)
	Stop()
}
