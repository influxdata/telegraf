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

// deepcopy returns a deep copy of the BatchPoints object. This is primarily so
// we can do multithreaded output flushing (see Agent.flush)
func (bp *BatchPoints) deepcopy() *BatchPoints {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	var bpc BatchPoints
	bpc.Time = bp.Time
	bpc.Precision = bp.Precision

	bpc.Tags = make(map[string]string)
	for k, v := range bp.Tags {
		bpc.Tags[k] = v
	}

	var pts []client.Point
	for _, pt := range bp.Points {
		var ptc client.Point

		ptc.Measurement = pt.Measurement
		ptc.Time = pt.Time
		ptc.Precision = pt.Precision
		ptc.Raw = pt.Raw

		ptc.Tags = make(map[string]string)
		ptc.Fields = make(map[string]interface{})

		for k, v := range pt.Tags {
			ptc.Tags[k] = v
		}

		for k, v := range pt.Fields {
			ptc.Fields[k] = v
		}
		pts = append(pts, ptc)
	}

	bpc.Points = pts
	return &bpc
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

// AddFields will eventually replace the Add function, once we move to having a
// single plugin as a single measurement with multiple fields
func (bp *BatchPoints) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
) {
	bp.AddFieldsWithTime(measurement, fields, tags, time.Time{})
}

// AddFieldsWithTime adds a measurement with a provided timestamp
func (bp *BatchPoints) AddFieldsWithTime(
	measurement string,
	fields map[string]interface{},
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

		for k, v := range fields {
			vals = append(vals, fmt.Sprintf("%s=%v", k, v))
		}

		sort.Strings(tg)
		sort.Strings(vals)

		fmt.Printf("> [%s] %s %s\n", strings.Join(tg, " "), measurement, strings.Join(vals, " "))
	}

	bp.Points = append(bp.Points, client.Point{
		Measurement: measurement,
		Tags:        tags,
		Fields:      fields,
		Time:        timestamp,
	})
}
