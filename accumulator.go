package telegraf

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/influxdb/influxdb/client"
)

type BatchPoints struct {
	client.BatchPoints

	Debug bool

	Prefix string

	Config *ConfiguredPlugin
}

func (bp *BatchPoints) Add(name string, val interface{}, tags map[string]string) {
	name = bp.Prefix + name

	if bp.Config != nil {
		if !bp.Config.ShouldPass(name) {
			return
		}
	}

	if bp.Debug {
		var tg []string

		for k, v := range tags {
			tg = append(tg, fmt.Sprintf("%s=\"%s\"", k, v))
		}

		sort.Strings(tg)

		fmt.Printf("> [%s] %s value=%v\n", strings.Join(tg, " "), name, val)
	}

	bp.Points = append(bp.Points, client.Point{
		Name: name,
		Tags: tags,
		Fields: map[string]interface{}{
			"value": val,
		},
	})
}

func (bp *BatchPoints) AddValuesWithTime(
	name string,
	values map[string]interface{},
	tags map[string]string,
	timestamp time.Time,
) {
	name = bp.Prefix + name

	if bp.Config != nil {
		if !bp.Config.ShouldPass(name) {
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

		fmt.Printf("> [%s] %s %s\n", strings.Join(tg, " "), name, strings.Join(vals, " "))
	}

	bp.Points = append(bp.Points, client.Point{
		Name:   name,
		Tags:   tags,
		Fields: values,
		Time:   timestamp,
	})
}
