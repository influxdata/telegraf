package telegraf

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdb/influxdb/client"
)

// BatchPoints is used to send a batch of data in a single write from telegraf
// to influx
type BatchPoints struct {
	mu sync.Mutex

	client.BatchPoints

	Debug bool

	Prefix string

	Config *ConfiguredPlugin
}

// Add adds a measurement
func (bp *BatchPoints) Add(
	measurement string,
	val interface{},
	tags map[string]string,
) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	measurement = bp.Prefix + measurement

	if bp.Config != nil {
		if !bp.Config.ShouldPass(measurement, tags) {
			return
		}
	}

	if bp.Debug {
		var tg []string

		for k, v := range tags {
			tg = append(tg, fmt.Sprintf("%s=\"%s\"", k, v))
		}

		sort.Strings(tg)

		fmt.Printf("> [%s] %s value=%v\n", strings.Join(tg, " "), measurement, val)
	}

	bp.Points = append(bp.Points, client.Point{
		Measurement: measurement,
		Tags:        tags,
		Fields: map[string]interface{}{
			"value": val,
		},
	})
}

// AddValuesWithTime adds a measurement with a provided timestamp
func (bp *BatchPoints) AddValuesWithTime(
	measurement string,
	values map[string]interface{},
	tags map[string]string,
	timestamp time.Time,
) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	measurement = bp.Prefix + measurement

	if bp.Config != nil {
		if !bp.Config.ShouldPass(measurement, tags) {
			return
		}
	}

	if bp.Debug {
		var tg []string

		for k, v := range tags {
			tg = append(tg, fmt.Sprintf("%s=\"%s\"", k, v))
		}

		var vals []string

		for k, v := range values {
			vals = append(vals, fmt.Sprintf("%s=%v", k, v))
		}

		sort.Strings(tg)
		sort.Strings(vals)

		fmt.Printf("> [%s] %s %s\n", strings.Join(tg, " "), measurement, strings.Join(vals, " "))
	}

	bp.Points = append(bp.Points, client.Point{
		Measurement: measurement,
		Tags:        tags,
		Fields:      values,
		Time:        timestamp,
	})
}
