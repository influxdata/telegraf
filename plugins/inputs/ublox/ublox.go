//go:generate ../../../tools/readme_config_includer/generator
package ublox

import (
	_ "embed"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type UbloxDataCollector struct {
	UbloxPTY string          `toml:"ublox_pty"`
	Log      telegraf.Logger `toml:"-"`
	posCh    chan GPSPos
	tdCh     chan int64
	errCh    chan error
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
	s.posCh = make(chan GPSPos, 1)
	s.tdCh = make(chan int64, 1)
	s.errCh = make(chan error, 1)
	go func() {
		reader := NewUbloxReader(s.UbloxPTY)
		lastFusionMode := None
		for {
			pos, err := reader.Pop(true)
			if err != nil {
				select {
				case s.errCh <- err:
				default:
				}
				continue
			} else if pos == nil {
				time.Sleep(time.Second * 1)
				continue
			}

			// aggregate fusion mode
			if pos.FusionMode == None {
				pos.FusionMode = lastFusionMode
			} else {
				lastFusionMode = pos.FusionMode
			}

			if pos.Active {
				now := time.Now()
				select {
				case s.tdCh <- now.Sub(pos.Ts).Milliseconds():
				default:
				}
			}

			select {
			case s.posCh <- *pos:
			default:
			}
		}
	}()
	return nil
}

func (s *UbloxDataCollector) Gather(acc telegraf.Accumulator) error {
	select {
	case lastPos := <-s.posCh:
		metrics := make(map[string]interface{})
		metrics["active"] = lastPos.Active
		metrics["lon"] = lastPos.Lon
		metrics["lat"] = lastPos.Lat
		metrics["heading"] = lastPos.Heading
		metrics["pdop"] = lastPos.Pdop

		if lastPos.FusionMode != None {
			metrics["fusion_mode"] = lastPos.FusionMode
		}

		select {
		case timeDiff := <-s.tdCh:
			metrics["system_gps_time_diff_ms"] = timeDiff
		default:
		}

		acc.AddFields("ublox-data", metrics, nil)
	case err := <-s.errCh:
		return err
	default:
	}

	return nil
}

func init() {
	inputs.Add("ublox", func() telegraf.Input { return &UbloxDataCollector{} })
}
