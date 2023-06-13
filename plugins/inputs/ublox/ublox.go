//go:generate ../../../tools/readme_config_includer/generator
package ublox

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type UbloxDataCollector struct {
	UbloxPTY string          `toml:"ublox_pty"`
	Log      telegraf.Logger `toml:"-"`
	reader   *UbloxReader
}

func (*UbloxDataCollector) Description() string {
	return "Read ublox metrics"
}

func (*UbloxDataCollector) SampleConfig() string {
	return `
[[inputs.ublox]]
    ublox_pty = "/tmp/ptyGPSRO_tlg"
`
}

// Init is for setup, and validating config.
func (s *UbloxDataCollector) Init() error {
	s.reader = NewUbloxReader(s.UbloxPTY)
	return nil
}

func (s *UbloxDataCollector) Gather(acc telegraf.Accumulator) error {
	var lastPos *GPSPos

	// read all buffered messages and return last one
	for {
		pos, err := s.reader.Pop(false)
		if err != nil {
			return err
		} else if pos == nil {
			break
		}

		lastFusionMode := lastPos.FusionMode

		lastPos = pos

		// aggregate fusion mode
		if pos.FusionMode == None {
			lastPos.FusionMode = lastFusionMode
		}
	}

	if lastPos != nil {
		metrics := make(map[string]interface{})
		metrics["active"] = lastPos.Active
		metrics["lon"] = lastPos.Lon
		metrics["lat"] = lastPos.Lat
		metrics["heading"] = lastPos.Heading
		metrics["pdop"] = lastPos.Pdop

		if lastPos.FusionMode != None {
			metrics["fusion_mode"] = lastPos.FusionMode
		}

		acc.AddFields("ublox-data", metrics, nil)
	}

	return nil
}

func init() {
	inputs.Add("ublox", func() telegraf.Input { return &UbloxDataCollector{} })
}
