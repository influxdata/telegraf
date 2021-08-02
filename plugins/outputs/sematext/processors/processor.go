package processors

import "github.com/influxdata/telegraf"

// MetricProcessor is interface that should be implemented by modules which adjust Telegraf metrics to match Sematext
// format.
type MetricProcessor interface {
	// Process makes adjustments to a single metric instance to be compliant with Sematext backend
	Process(metric telegraf.Metric) error

	Close()
}

// BatchProcessor is used to execute actions on the level of a whole batch of metrics. Batch processors are run before
// any Metric processors kick in, so metrics produced by a batch processor can count on further being decorated by
// metric processors. Also, metrics batch processors work on are "original" Telegraf metrics without any Sematext
// specific adjustments.
type BatchProcessor interface {
	Process(metrics []telegraf.Metric) ([]telegraf.Metric, error)

	Close()
}
