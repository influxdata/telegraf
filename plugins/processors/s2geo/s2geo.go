package geo

import (
	"fmt"

	"github.com/golang/geo/s2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Geo struct {
	LatField  string `toml:"lat_field"`
	LonField  string `toml:"lon_field"`
	TagKey    string `toml:"tag_key"`
	CellLevel int    `toml:"cell_level"`
}

var SampleConfig = `
  ## The name of the lat and lon fields containing WGS-84 latitude and
  ## longitude in decimal degrees.
  # lat_field = "lat"
  # lon_field = "lon"

  ## New tag to create
  # tag_key = "s2_cell_id"

  ## Cell level (see https://s2geometry.io/resources/s2cell_statistics.html)
  # cell_level = 9
`

func (g *Geo) SampleConfig() string {
	return SampleConfig
}

func (g *Geo) Description() string {
	return "Add the S2 Cell ID as a tag based on latitude and longitude fields"
}

func (g *Geo) Init() error {
	if g.CellLevel < 0 || g.CellLevel > 30 {
		return fmt.Errorf("invalid cell level %d", g.CellLevel)
	}
	return nil
}

func (g *Geo) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, point := range in {
		var latOk, lonOk bool
		var lat, lon float64
		for _, field := range point.FieldList() {
			switch field.Key {
			case g.LatField:
				lat, latOk = field.Value.(float64)
			case g.LonField:
				lon, lonOk = field.Value.(float64)
			}
		}
		if latOk && lonOk {
			cellID := s2.CellIDFromLatLng(s2.LatLngFromDegrees(lat, lon))
			if cellID.IsValid() {
				value := cellID.Parent(g.CellLevel).ToToken()
				point.AddTag(g.TagKey, value)
			}
		}
	}
	return in
}

func init() {
	processors.Add("s2geo", func() telegraf.Processor {
		return &Geo{
			LatField:  "lat",
			LonField:  "lon",
			TagKey:    "s2_cell_id",
			CellLevel: 9,
		}
	})
}
