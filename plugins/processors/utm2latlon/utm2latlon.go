package utm2latlon

import (
	"log"

	"github.com/im7mortal/UTM"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
[[processors.geohash]]
`

type MineGrid struct {
	K1_A          float64
	K2_B          float64
	X0_C          float64
	K3_D          float64
	K4_E          float64
	Y0_F          float64
	FalseEasting  float64
	FalseNorthing float64
}

type Utm2LatLon struct {
	X          string
	Y          string
	Latitude   string
	Longitude  string
	ZoneNumber int
	ZoneLetter string
	MineGrid   MineGrid

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

func (m *MineGrid) ConvertToEastingNorthing(x, y float64) (easting float64, northing float64, err error) {
	e := (x * m.K1_A) + (y * m.K2_B) + m.X0_C - m.FalseEasting
	n := (x * m.K3_D) + (y * m.K4_E) + m.Y0_F - m.FalseNorthing
	return e, n, nil
}

func (p *Utm2LatLon) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if !p.configParsed {
		p.configParsed = p.ParseConfig()
	}

	var x, y telegraf.Metric
	for _, m := range in {
		if _, xOk := m.Fields()[p.X]; xOk {
			x = m
		}

		if _, yOk := m.Fields()[p.Y]; yOk {
			y = m
		}
	}

	if x != nil && y != nil && x.Time() == y.Time() {
		fields := make(map[string]interface{})
		xval := x.Fields()[p.X].(float64)
		yval := y.Fields()[p.Y].(float64)

		utmEasting, utmNorthing, err := p.MineGrid.ConvertToEastingNorthing(xval, yval)
		if err != nil {
			log.Println("ERROR [UTM.Convert]: Could not convert to UTM", x, y, "ERROR:", err)
		}
		latitude, longitude, err := UTM.ToLatLon(utmEasting, utmNorthing, p.ZoneNumber, p.ZoneLetter)
		if err != nil {
			log.Println("ERROR [UTM.Convert]: Could not convert x/y", utmEasting, utmNorthing, "ERROR:", err)
		}
		fields[p.Latitude] = latitude
		fields[p.Longitude] = longitude
		newMetric, err := metric.New(x.Name(), x.Tags(), fields, x.Time())
		if err != nil {
			log.Println("ERROR [UTM.Apply]: Could not create a new metric")
		}
		in = append(in, newMetric)
	}

	return in
}

func init() {
	processors.Add("utm2latlon", func() telegraf.Processor {
		return &Utm2LatLon{}
	})
}
