package buffer

import "github.com/influxdata/telegraf"

type Buffer interface {
	IsEmpty() bool
	Len() int
	Add(...telegraf.Metric)
	Batch(int) []telegraf.Metric
}
