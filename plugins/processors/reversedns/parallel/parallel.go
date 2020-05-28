package parallel

import "github.com/influxdata/telegraf"

type Parallel interface {
	Do(fn func(acc telegraf.MetricStreamAccumulator))
	Wait()
}
