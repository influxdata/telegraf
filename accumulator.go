package telegraf

import (
	"fmt"
	"sync"
	"time"

	"github.com/influxdb/influxdb/client/v2"
)

type Accumulator interface {
	Add(measurement string, value interface{},
		tags map[string]string, t ...time.Time)
	AddFields(measurement string, fields map[string]interface{},
		tags map[string]string, t ...time.Time)

	SetDefaultTags(tags map[string]string)
	AddDefaultTag(key, value string)

	Prefix() string
	SetPrefix(prefix string)

	Debug() bool
	SetDebug(enabled bool)
}

func NewAccumulator(
	plugin *ConfiguredPlugin,
	points chan *client.Point,
) Accumulator {
	acc := accumulator{}
	acc.points = points
	acc.plugin = plugin
	return &acc
}

type accumulator struct {
	sync.Mutex

	points chan *client.Point

	defaultTags map[string]string

	debug bool

	plugin *ConfiguredPlugin

	prefix string
}

func (ac *accumulator) Add(
	measurement string,
	value interface{},
	tags map[string]string,
	t ...time.Time,
) {
	fields := make(map[string]interface{})
	fields["value"] = value
	ac.AddFields(measurement, fields, tags, t...)
}

func (ac *accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {

	if tags == nil {
		tags = make(map[string]string)
	}

	// InfluxDB client/points does not support writing uint64
	// TODO fix when it does
	// https://github.com/influxdb/influxdb/pull/4508
	for k, v := range fields {
		switch val := v.(type) {
		case uint64:
			if val < uint64(9223372036854775808) {
				fields[k] = int64(val)
			} else {
				fields[k] = int64(9223372036854775807)
			}
		}
	}

	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}

	if ac.plugin != nil {
		if !ac.plugin.ShouldPass(measurement, tags) {
			return
		}
	}

	for k, v := range ac.defaultTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	if ac.prefix != "" {
		measurement = ac.prefix + measurement
	}

	pt := client.NewPoint(measurement, tags, fields, timestamp)
	if ac.debug {
		fmt.Println("> " + pt.String())
	}
	ac.points <- pt
}

func (ac *accumulator) SetDefaultTags(tags map[string]string) {
	ac.defaultTags = tags
}

func (ac *accumulator) AddDefaultTag(key, value string) {
	ac.defaultTags[key] = value
}

func (ac *accumulator) Prefix() string {
	return ac.prefix
}

func (ac *accumulator) SetPrefix(prefix string) {
	ac.prefix = prefix
}

func (ac *accumulator) Debug() bool {
	return ac.debug
}

func (ac *accumulator) SetDebug(debug bool) {
	ac.debug = debug
}
