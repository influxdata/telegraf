package http

import (
	"time"

	"github.com/influxdata/telegraf"
	// "github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	// "github.com/influxdata/telegraf/plugins/serializers/influx"
	// "github.com/stretchr/testify/require"
)

func getMetric() telegraf.Metric {
	m, err := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	if err != nil {
		panic(err)
	}
	return m
}
