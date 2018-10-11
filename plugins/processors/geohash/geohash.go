package geohash

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/mmcloughlin/geohash"
	"github.com/srclosson/telegraf/metric"
)

var sampleConfig = `
[[processors.geohash]]
`

type GeoHash struct {
	Longitude string
	Latitude  string
	Name      string

	configParsed bool
	lat          telegraf.Metric
	lng          telegraf.Metric
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

	for _, m := range in {
		if _, latOk := m.Fields()[p.Latitude]; latOk {
			p.lat = m
		}

		if _, lngOk := m.Fields()[p.Longitude]; lngOk {
			p.lng = m
		}
	}

	if p.lat != nil && p.lng != nil && p.lat.Time() == p.lng.Time() {
		fields := make(map[string]interface{})
		lat := p.lat.Fields()[p.Latitude].(float64)
		lng := p.lng.Fields()[p.Longitude].(float64)
		fields[p.Name] = geohash.Encode(lat, lng)
		ghMetric, err := metric.New(p.lat.Name(), p.lat.Tags(), fields, p.lat.Time())
		if err != nil {
			log.Println("ERROR [geohash.Apply]: Could not create a new metric")
		}
		in = append(in, ghMetric)

		/* We don't need these anymore */
		p.lat = nil
		p.lng = nil
	}

	return in
}

func init() {
	processors.Add("geohash", func() telegraf.Processor {
		return &GeoHash{}
	})
}
