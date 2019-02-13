package utm2latlon

import (
	"log"

	"github.com/im7mortal/UTM"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
[[processors.geohash]]
`

type Utm2LatLon struct {
	Easting    string
	Northing   string
	Latitude   string
	Longitude  string
	ZoneNumber int
	ZoneLetter string

	configParsed bool
}

func (p *Utm2LatLon) SampleConfig() string {
	return sampleConfig
}

func (p *Utm2LatLon) Description() string {
	return "Transform UTM x/y/z into latitude and longitude"
}

func (p *Utm2LatLon) ParseConfig() bool {
	return true
}

func (p *Utm2LatLon) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if !p.configParsed {
		p.configParsed = p.ParseConfig()
	}

	var metrics []telegraf.Metric
	for _, m := range in {
		if m.HasField(p.Easting) && m.HasField(p.Northing) {
			metrics = append(metrics, m)
		}
	}

	for _, m := range metrics {
		utmEasting := m.Fields()[p.Easting].(float64)
		utmNorthing := m.Fields()[p.Northing].(float64)
		latitude, longitude, err := UTM.ToLatLon(utmEasting, utmNorthing, p.ZoneNumber, p.ZoneLetter)
		if err != nil {
			log.Println("ERROR [UTM.Convert]: Could not convert easting/northing", utmEasting, utmNorthing, "ERROR:", err)
			continue
		}
		m.AddField(p.Latitude, latitude)
		m.AddField(p.Longitude, longitude)
	}

	return in
}

func init() {
	processors.Add("utm2latlon", func() telegraf.Processor {
		return &Utm2LatLon{}
	})
}
