package apcupsd

import (
	"context"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/mdlayher/apcupsd"
)

const defaultAddress = "127.0.0.1:3551"

var defaultTimeout = internal.Duration{Duration: time.Duration(time.Second * 5)}

type ApcUpsd struct {
	Servers []string
	Timeout internal.Duration
}

func (_ *ApcUpsd) Description() string {
	return "Monitor APC UPSes connected to apcupsd"
}

var sampleConfig = `
  # a list of running apcupsd server to connect to. 
  # If not provided will default to 127.0.0.1:3551
  servers = ["127.0.0.1:3551"]
  timeout = "5s"
`

func (_ *ApcUpsd) SampleConfig() string {
	return sampleConfig
}

func (h *ApcUpsd) Gather(acc telegraf.Accumulator) error {
	for _, addr := range h.Servers {
		status, err := fetchStatus(addr, h.Timeout.Duration)
		if err != nil {
			return err
		}

		tags := map[string]string{
			"serial":   status.SerialNumber,
			"ups_name": status.UPSName,
			"online":   status.Status,
		}

		fields := map[string]interface{}{
			"online":                 boolToInt(status.Status == "ONLINE"),
			"input_voltage":          status.LineVoltage,
			"load_percent":           status.LoadPercent,
			"battery_charge_percent": status.BatteryChargePercent,
			"time_left_minutes":      status.TimeLeft.Minutes(),
			"output_voltage":         status.OutputVoltage,
			"internal_temp":          status.InternalTemp,
			"battery_voltage":        status.BatteryVoltage,
			"input_frequency":        status.LineFrequency,
			"time_on_battery":        status.TimeOnBattery.Minutes(),
		}

		acc.AddFields("apcupsd", fields, tags)
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func fetchStatus(addr string, timeout time.Duration) (*apcupsd.Status, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client, err := apcupsd.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	return client.Status()
}

func init() {
	inputs.Add("apcupsd", func() telegraf.Input {
		return &ApcUpsd{
			Servers: []string{defaultAddress},
			Timeout: defaultTimeout,
		}
	})
}
