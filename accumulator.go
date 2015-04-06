package tivan

import "github.com/influxdb/influxdb/client"

type BatchPoints struct {
	client.BatchPoints
}

func (bp *BatchPoints) Add(name string, val interface{}, tags map[string]string) {
	bp.Points = append(bp.Points, client.Point{
		Name: name,
		Tags: tags,
		Fields: map[string]interface{}{
			"value": val,
		},
	})
}
