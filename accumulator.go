package tivan

import (
	"fmt"
	"sort"
	"strings"

	"github.com/influxdb/influxdb/client"
)

type BatchPoints struct {
	client.BatchPoints

	Debug bool
}

func (bp *BatchPoints) Add(name string, val interface{}, tags map[string]string) {
	if bp.Debug {
		var tg []string

		for k, v := range tags {
			tg = append(tg, fmt.Sprintf("%s=\"%s\"", k, v))
		}

		sort.Strings(tg)

		fmt.Printf("> [%s] %s=%v\n", strings.Join(tg, " "), name, val)
	}

	bp.Points = append(bp.Points, client.Point{
		Name: name,
		Tags: tags,
		Fields: map[string]interface{}{
			"value": val,
		},
	})
}
