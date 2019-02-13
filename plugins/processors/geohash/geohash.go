package geohash

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/mmcloughlin/geohash"
)

var sampleConfig = `
[[processors.geohash]]
`

type GeoHash struct {
	Longitude string
	Latitude  string
	Name      string

	configParsed bool
}

func (p *GeoHash) SampleConfig() string {
	return sampleConfig
}

func (p *GeoHash) Description() string {
	return "Transform longitude and Latitude metrics into a geohash"
}

func (p *GeoHash) ParseConfig() bool {
	return true
}

func (p *GeoHash) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if !p.configParsed {
		p.configParsed = p.ParseConfig()
	}

	var metrics []telegraf.Metric
	for _, m := range in {
		if m.HasField(p.Longitude) && m.HasField(p.Latitude) {
			metrics = append(metrics, m)
		}
	}

	for _, m := range metrics {
		lat := m.Fields()[p.Latitude].(float64)
		lng := m.Fields()[p.Longitude].(float64)
		gh := geohash.Encode(lat, lng)
		m.AddField(p.Name, gh)
	}

	return in
}

func init() {
	processors.Add("geohash", func() telegraf.Processor {
		return &GeoHash{}
	})
}
