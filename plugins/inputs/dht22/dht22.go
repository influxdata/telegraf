// +build linux
package dht22

import (
	"github.com/gdunstone/go-dht"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"math"
)

var DHTConfig = `
  ## Set the GPIO pin
  pin = 14
  ## how many times to retry for a good reading.
  retry = 10
  ## Additionally calculate Vapor Pressure Deficit in kPa
  calcvpd = true
  ## divisor/multiplier for VPD (1000 to transform to Pa)
  vpdmultiplier = 1
`

type DHT struct {
	Pin           int
	Retry         int
	CalcVpd       bool
	VpdMultiplier float64
}

func (*DHT) Description() string {
	return "Monitor DHT22 connected to GPIO"
}

func (*DHT) SampleConfig() string {
	return DHTConfig
}

func (s *DHT) Gather(acc telegraf.Accumulator) error {
	temperature, humidity, retries, err := dht.ReadDHTxxWithRetry(dht.DHT22, s.Pin, s.Retry)
	if err != nil {
		return err
	}

	fields := make(map[string]interface{})

	if s.CalcVpd {
		// calculate vpd
		// J. Win. (https://physics.stackexchange.com/users/1680/j-win),
		// How can I calculate Vapor Pressure Deficit from Temperature and Relative Humidity?,
		// URL (version: 2011-02-03): https://physics.stackexchange.com/q/4553
		temperature64 := float64(temperature)
		humidity64 := float64(humidity)

		es := 0.6108 * math.Exp(17.27*temperature64/(temperature64+237.3))
		ea := humidity64 / 100 * es

		// this equation returns a negative value, which while technically correct,
		// is invalid in this case because we are talking about a deficit.
		vpd := (ea - es) * -1
		fields["vpd"] = vpd * s.VpdMultiplier
	}

	fields["temperature"] = temperature
	fields["humidity"] = humidity
	fields["retries"] = retries

	acc.AddFields("dht22", fields, nil)

	return nil
}

func init() {
	inputs.Add("dht22", func() telegraf.Input {
		return &DHT{
			14,
			10,
			true,
			1.0,
		}
	})
}
